package screen

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"financli/internal/application/usecase"
	"financli/internal/domain/entity"
	"financli/internal/interfaces/tui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type BillsModel struct {
	ctx         context.Context
	billUseCase *usecase.BillUseCase

	// Data
	bills []*entity.Bill

	// View state
	selectedIndex int
	viewMode      BillViewMode

	// Loading and errors
	loading bool
	err     error

	// Form state
	formModel *BillFormModel

	// Payment state
	paymentModel *BillPaymentFormModel

	// Confirmation state
	showConfirmDelete bool
	confirmMessage    string

	// Window dimensions
	width  int
	height int
}

type BillViewMode int

const (
	BillViewList BillViewMode = iota
	BillViewForm
	BillViewDetails
	BillViewPayment
	BillViewConfirm
)

type BillFormModel struct {
	// Form fields
	name        string
	description string
	totalAmount string
	startDate   string
	endDate     string
	dueDate     string

	// Input fields
	nameInput        string
	descriptionInput string
	amountInput      string
	startDateInput   string
	endDateInput     string
	dueDateInput     string

	// Navigation
	focusedField int

	// Edit state
	editing   bool
	editingID *uuid.UUID
}

type BillPaymentFormModel struct {
	billID uuid.UUID
	bill   *entity.Bill

	// Payment amount
	amountInput string

	// Navigation
	focusedField int
}

// Message types
type billsLoadedMsg struct {
	bills []*entity.Bill
}

type billActionMsg struct{}

func NewBillsModel(ctx context.Context, billUC *usecase.BillUseCase) tea.Model {
	return &BillsModel{
		ctx:         ctx,
		billUseCase: billUC,
		viewMode:    BillViewList,
		loading:     true,
		formModel:   &BillFormModel{},
	}
}

func (m *BillsModel) Init() tea.Cmd {
	return m.loadBills
}

func (m *BillsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case billsLoadedMsg:
		m.loading = false
		m.bills = msg.bills
		if len(m.bills) > 0 && m.selectedIndex >= len(m.bills) {
			m.selectedIndex = len(m.bills) - 1
		}
		return m, nil

	case billActionMsg:
		m.loading = false
		m.viewMode = BillViewList
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
		return m, m.loadBills

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch m.viewMode {
		case BillViewList:
			return m.handleListKeys(msg)
		case BillViewForm:
			return m.handleFormKeys(msg)
		case BillViewDetails:
			return m.handleDetailsKeys(msg)
		case BillViewPayment:
			return m.handlePaymentKeys(msg)
		case BillViewConfirm:
			return m.handleConfirmKeys(msg)
		}
	}

	return m, nil
}

func (m *BillsModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.bills)-1 {
			m.selectedIndex++
		}
	case "enter":
		if len(m.bills) > 0 {
			m.viewMode = BillViewDetails
		}
	case "n":
		m.viewMode = BillViewForm
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
	case "e":
		if len(m.bills) > 0 {
			return m.editBill()
		}
	case "p":
		if len(m.bills) > 0 {
			return m.startPayment()
		}
	case "c":
		if len(m.bills) > 0 {
			return m.closeBill()
		}
	case "d":
		if len(m.bills) > 0 {
			m.viewMode = BillViewConfirm
			m.showConfirmDelete = true
		}
	case "r":
		m.loading = true
		return m, m.loadBills
	case "b":
		return m, func() tea.Msg { return BackToDashboardMsg{} }
	}

	return m, nil
}

func (m *BillsModel) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = BillViewList
		m.resetForm()
	case "tab", "down":
		m.formModel.focusedField = (m.formModel.focusedField + 1) % 8
	case "shift+tab", "up":
		m.formModel.focusedField = (m.formModel.focusedField - 1 + 8) % 8
	case "enter":
		if m.formModel.focusedField == 6 {
			return m.submitForm()
		} else if m.formModel.focusedField == 7 {
			// Cancel button
			m.viewMode = BillViewList
			m.resetForm()
		}
	default:
		return m.handleFormInput(msg)
	}

	return m, nil
}

func (m *BillsModel) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only handle input for fields 0-5 (form fields)
	// Fields 6-7 are buttons
	if m.formModel.focusedField > 5 {
		return m, nil
	}

	switch m.formModel.focusedField {
	case 0: // Name
		switch msg.String() {
		case "backspace":
			if len(m.formModel.nameInput) > 0 {
				m.formModel.nameInput = m.formModel.nameInput[:len(m.formModel.nameInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.formModel.nameInput += msg.String()
			}
		}
	case 1: // Description
		switch msg.String() {
		case "backspace":
			if len(m.formModel.descriptionInput) > 0 {
				m.formModel.descriptionInput = m.formModel.descriptionInput[:len(m.formModel.descriptionInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.formModel.descriptionInput += msg.String()
			}
		}
	case 2: // Total Amount
		switch msg.String() {
		case "backspace":
			if len(m.formModel.amountInput) > 0 {
				m.formModel.amountInput = m.formModel.amountInput[:len(m.formModel.amountInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == ".") {
				m.formModel.amountInput += msg.String()
			}
		}
	case 3: // Start Date
		switch msg.String() {
		case "backspace":
			if len(m.formModel.startDateInput) > 0 {
				m.formModel.startDateInput = m.formModel.startDateInput[:len(m.formModel.startDateInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == "-") {
				m.formModel.startDateInput += msg.String()
			}
		}
	case 4: // End Date
		switch msg.String() {
		case "backspace":
			if len(m.formModel.endDateInput) > 0 {
				m.formModel.endDateInput = m.formModel.endDateInput[:len(m.formModel.endDateInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == "-") {
				m.formModel.endDateInput += msg.String()
			}
		}
	case 5: // Due Date
		switch msg.String() {
		case "backspace":
			if len(m.formModel.dueDateInput) > 0 {
				m.formModel.dueDateInput = m.formModel.dueDateInput[:len(m.formModel.dueDateInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == "-") {
				m.formModel.dueDateInput += msg.String()
			}
		}
	}

	return m, nil
}

func (m *BillsModel) handleDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "b":
		m.viewMode = BillViewList
	case "e":
		return m.editBill()
	case "p":
		return m.startPayment()
	case "c":
		return m.closeBill()
	case "d":
		m.viewMode = BillViewConfirm
		m.showConfirmDelete = true
	}

	return m, nil
}

func (m *BillsModel) handlePaymentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = BillViewList
		m.paymentModel = nil
	case "tab", "down":
		m.paymentModel.focusedField = (m.paymentModel.focusedField + 1) % 3
	case "shift+tab", "up":
		m.paymentModel.focusedField = (m.paymentModel.focusedField - 1 + 3) % 3
	case "enter":
		if m.paymentModel.focusedField == 1 {
			return m.submitPayment()
		} else if m.paymentModel.focusedField == 2 {
			// Cancel button
			m.viewMode = BillViewList
			m.paymentModel = nil
		}
	default:
		if m.paymentModel.focusedField == 0 {
			switch msg.String() {
			case "backspace":
				if len(m.paymentModel.amountInput) > 0 {
					m.paymentModel.amountInput = m.paymentModel.amountInput[:len(m.paymentModel.amountInput)-1]
				}
			default:
				if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == ".") {
					m.paymentModel.amountInput += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m *BillsModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.viewMode = BillViewList
		m.showConfirmDelete = false
		m.loading = true
		return m, m.deleteBill
	case "n", "esc":
		m.viewMode = BillViewList
		m.showConfirmDelete = false
	}

	return m, nil
}

func (m *BillsModel) View() string {
	if m.loading {
		return style.InfoStyle.Render("Loading bills...")
	}

	if m.err != nil {
		return style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.viewMode {
	case BillViewList:
		return m.renderBillsList()
	case BillViewForm:
		return m.renderBillForm()
	case BillViewDetails:
		return m.renderBillDetails()
	case BillViewPayment:
		return m.renderPaymentForm()
	case BillViewConfirm:
		return m.renderConfirmDialog()
	}

	return ""
}

func (m *BillsModel) renderBillsList() string {
	var sections []string

	title := style.TitleStyle.Render("📋 Bills Management")
	sections = append(sections, title)

	if len(m.bills) == 0 {
		empty := style.InfoStyle.Render("No bills found. Press 'n' to create your first bill.")
		sections = append(sections, empty)
	} else {
		table := m.renderBillsTable()
		sections = append(sections, table)

		summary := m.renderBillsSummary()
		sections = append(sections, summary)
	}

	help := m.renderListHelp()
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *BillsModel) renderBillsTable() string {
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)

	headers := []string{"Status", "Bill Name", "Total", "Paid", "Progress", "Due Date"}
	headerRow := style.TableHeaderStyle.Render(
		fmt.Sprintf("%-8s %-20s %-12s %-12s %-20s %s",
			headers[0], headers[1], headers[2], headers[3], headers[4], headers[5]),
	)

	var rows []string
	rows = append(rows, headerRow)

	for i, bill := range m.bills {
		status := m.getBillStatusIcon(bill.Status)
		name := truncateString(bill.Name, 20)
		total := bill.TotalAmount.String()
		paid := bill.PaidAmount.String()
		progress := m.renderProgressBar(bill.GetPaymentPercentage(), 20)
		dueDate := bill.DueDate.Format("2006-01-02")

		// Color code due date if overdue
		if bill.Status == entity.BillStatusOverdue {
			dueDate = style.ErrorStyle.Render(dueDate)
		}

		row := fmt.Sprintf("%-8s %-20s %-12s %-12s %s %s",
			status, name, total, paid, progress, dueDate)

		if i == m.selectedIndex {
			row = style.SelectedMenuItemStyle.Render("► " + row)
		} else {
			row = style.MenuItemStyle.Render("  " + row)
		}

		rows = append(rows, row)
	}

	table := strings.Join(rows, "\n")
	return tableStyle.Render(table)
}

func (m *BillsModel) renderBillsSummary() string {
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)

	totalBills := len(m.bills)
	openBills := 0
	overdueBills := 0
	totalAmount := 0.0
	totalPaid := 0.0

	for _, bill := range m.bills {
		if bill.Status == entity.BillStatusOpen {
			openBills++
		} else if bill.Status == entity.BillStatusOverdue {
			overdueBills++
		}
		totalAmount += bill.TotalAmount.Amount()
		totalPaid += bill.PaidAmount.Amount()
	}

	summary := []string{
		fmt.Sprintf("Total Bills: %d", totalBills),
		fmt.Sprintf("Open Bills: %d", openBills),
		fmt.Sprintf("Overdue Bills: %d", overdueBills),
		fmt.Sprintf("Total Amount: R$ %.2f", totalAmount),
		fmt.Sprintf("Total Paid: R$ %.2f", totalPaid),
		fmt.Sprintf("Remaining: R$ %.2f", totalAmount-totalPaid),
	}

	content := strings.Join(summary, " • ")
	return summaryStyle.Render(content)
}

func (m *BillsModel) renderBillDetails() string {
	if m.selectedIndex >= len(m.bills) {
		return ""
	}

	bill := m.bills[m.selectedIndex]

	var sections []string

	title := style.TitleStyle.Render(fmt.Sprintf("📋 Bill Details: %s", bill.Name))
	sections = append(sections, title)

	// Main details
	detailsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)

	details := []string{
		fmt.Sprintf("Status: %s %s", m.getBillStatusIcon(bill.Status), m.getBillStatusName(bill.Status)),
		fmt.Sprintf("Description: %s", bill.Description),
		"",
		fmt.Sprintf("Period: %s to %s", bill.StartDate.Format("2006-01-02"), bill.EndDate.Format("2006-01-02")),
		fmt.Sprintf("Due Date: %s", bill.DueDate.Format("2006-01-02")),
		"",
		fmt.Sprintf("Total Amount: %s", bill.TotalAmount.String()),
		fmt.Sprintf("Paid Amount: %s", bill.PaidAmount.String()),
	}

	remaining, _ := bill.GetRemainingAmount()
	details = append(details, fmt.Sprintf("Remaining: %s", remaining.String()))

	content := strings.Join(details, "\n")
	sections = append(sections, detailsStyle.Render(content))

	// Payment progress visualization
	progressStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Success).
		Padding(1, 2).
		MarginTop(1)

	percentage := bill.GetPaymentPercentage()
	progressBar := m.renderProgressBar(percentage, 50)
	progressInfo := fmt.Sprintf("Payment Progress: %.1f%%\n\n%s", percentage, progressBar)

	sections = append(sections, progressStyle.Render(progressInfo))

	// Actions help
	help := "[e] Edit • [p] Add Payment • [c] Close Bill • [d] Delete • [b] Back"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *BillsModel) renderBillForm() string {
	var sections []string

	title := "📝 New Bill"
	if m.formModel.editing {
		title = "✏️ Edit Bill"
	}
	sections = append(sections, style.TitleStyle.Render(title))

	form := m.renderForm()
	sections = append(sections, form)

	help := m.renderFormHelp()
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *BillsModel) renderForm() string {
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(2, 4).
		MarginTop(1)

	var fields []string

	fields = append(fields, m.renderFormField("Name:", m.formModel.nameInput, 0))
	fields = append(fields, m.renderFormField("Description:", m.formModel.descriptionInput, 1))
	fields = append(fields, m.renderFormField("Total Amount:", m.formModel.amountInput, 2))
	fields = append(fields, m.renderFormField("Start Date (YYYY-MM-DD):", m.formModel.startDateInput, 3))
	fields = append(fields, m.renderFormField("End Date (YYYY-MM-DD):", m.formModel.endDateInput, 4))
	fields = append(fields, m.renderFormField("Due Date (YYYY-MM-DD):", m.formModel.dueDateInput, 5))

	buttons := m.renderFormButtons()
	fields = append(fields, buttons)

	content := strings.Join(fields, "\n\n")
	return formStyle.Render(content)
}

func (m *BillsModel) renderFormField(label, value string, fieldIndex int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(25)

	var inputStyle lipgloss.Style
	if m.formModel.focusedField == fieldIndex {
		inputStyle = style.FocusedInputStyle.Width(30)
	} else {
		inputStyle = style.InputStyle.Width(30)
	}

	input := inputStyle.Render(value)

	if m.formModel.focusedField == fieldIndex {
		input = input + " ◄"
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label),
		input,
	)
}

func (m *BillsModel) renderFormButtons() string {
	submitText := "Create Bill"
	if m.formModel.editing {
		submitText = "Update Bill"
	}

	var submitStyle, cancelStyle lipgloss.Style

	// Submit button styling
	if m.formModel.focusedField == 6 {
		submitStyle = style.ButtonStyle.Background(style.Success)
	} else {
		submitStyle = style.SecondaryButtonStyle
	}

	// Cancel button styling
	if m.formModel.focusedField == 7 {
		cancelStyle = style.ButtonStyle.Background(style.Danger)
	} else {
		cancelStyle = style.SecondaryButtonStyle
	}

	submitBtn := submitStyle.Render(submitText)
	cancelBtn := cancelStyle.Render("Cancel")

	// Add focus indicators
	if m.formModel.focusedField == 6 {
		submitBtn = submitBtn + " ◄"
	} else if m.formModel.focusedField == 7 {
		cancelBtn = cancelBtn + " ◄"
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		submitBtn,
		lipgloss.NewStyle().MarginLeft(2).Render(cancelBtn),
	)
}

func (m *BillsModel) renderFormHelp() string {
	help := "[Tab] Next Field • [Shift+Tab] Previous • [Enter] Confirm • [Esc] Cancel"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

func (m *BillsModel) renderPaymentForm() string {
	if m.paymentModel == nil || m.paymentModel.bill == nil {
		return ""
	}

	var sections []string

	title := style.TitleStyle.Render(fmt.Sprintf("💰 Add Payment to: %s", m.paymentModel.bill.Name))
	sections = append(sections, title)

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Success).
		Padding(2, 4).
		MarginTop(1)

	var fields []string

	// Bill info
	fields = append(fields, style.InfoStyle.Render(fmt.Sprintf("Total Amount: %s", m.paymentModel.bill.TotalAmount.String())))
	fields = append(fields, style.InfoStyle.Render(fmt.Sprintf("Already Paid: %s", m.paymentModel.bill.PaidAmount.String())))

	remaining, _ := m.paymentModel.bill.GetRemainingAmount()
	fields = append(fields, style.WarningStyle.Render(fmt.Sprintf("Remaining: %s", remaining.String())))

	fields = append(fields, "") // Empty line

	// Payment amount input
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)

	var inputStyle lipgloss.Style
	if m.paymentModel.focusedField == 0 {
		inputStyle = style.FocusedInputStyle.Width(30)
	} else {
		inputStyle = style.InputStyle.Width(30)
	}

	input := inputStyle.Render(m.paymentModel.amountInput)
	if m.paymentModel.focusedField == 0 {
		input = input + " ◄"
	}

	paymentField := lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Payment Amount:"),
		input,
	)
	fields = append(fields, paymentField)

	// Buttons
	var submitStyle, cancelStyle lipgloss.Style

	if m.paymentModel.focusedField == 1 {
		submitStyle = style.ButtonStyle.Background(style.Success)
	} else {
		submitStyle = style.SecondaryButtonStyle
	}

	if m.paymentModel.focusedField == 2 {
		cancelStyle = style.ButtonStyle.Background(style.Danger)
	} else {
		cancelStyle = style.SecondaryButtonStyle
	}

	submitBtn := submitStyle.Render("Process Payment")
	cancelBtn := cancelStyle.Render("Cancel")

	if m.paymentModel.focusedField == 1 {
		submitBtn = submitBtn + " ◄"
	} else if m.paymentModel.focusedField == 2 {
		cancelBtn = cancelBtn + " ◄"
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		submitBtn,
		lipgloss.NewStyle().MarginLeft(2).Render(cancelBtn),
	)
	fields = append(fields, buttons)

	content := strings.Join(fields, "\n\n")
	sections = append(sections, formStyle.Render(content))

	help := "[Tab] Next Field • [Shift+Tab] Previous • [Enter] Confirm • [Esc] Cancel"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *BillsModel) renderConfirmDialog() string {
	if m.selectedIndex >= len(m.bills) {
		return ""
	}

	bill := m.bills[m.selectedIndex]

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Danger).
		Padding(2, 4).
		MarginTop(5)

	title := style.ErrorStyle.Render("⚠️  Confirm Delete")
	message := fmt.Sprintf("Are you sure you want to delete bill '%s'?", bill.Name)
	warning := style.WarningStyle.Render("This action cannot be undone!")
	help := "[y] Yes, Delete • [n] Cancel"

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		warning,
		"",
		help,
	)

	return dialogStyle.Render(content)
}

func (m *BillsModel) renderListHelp() string {
	help := "[↑/↓] Navigate • [Enter] View • [n] New • [e] Edit • [p] Payment • [c] Close • [d] Delete • [r] Refresh • [b] Back"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

func (m *BillsModel) renderProgressBar(percentage float64, width int) string {
	filled := int(percentage * float64(width) / 100)
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	color := m.getPaymentColor(percentage)
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

func (m *BillsModel) getPaymentColor(percentage float64) lipgloss.Color {
	if percentage >= 100 {
		return style.Success
	} else if percentage >= 50 {
		return style.Warning
	} else {
		return style.Danger
	}
}

func (m *BillsModel) getBillStatusIcon(status entity.BillStatus) string {
	switch status {
	case entity.BillStatusOpen:
		return "📄"
	case entity.BillStatusClosed:
		return "🔒"
	case entity.BillStatusPaid:
		return "✅"
	case entity.BillStatusOverdue:
		return "⚠️"
	default:
		return "❓"
	}
}

func (m *BillsModel) getBillStatusName(status entity.BillStatus) string {
	switch status {
	case entity.BillStatusOpen:
		return "Open"
	case entity.BillStatusClosed:
		return "Closed"
	case entity.BillStatusPaid:
		return "Paid"
	case entity.BillStatusOverdue:
		return "Overdue"
	default:
		return "Unknown"
	}
}

func (m *BillsModel) loadBills() tea.Msg {
	bills, err := m.billUseCase.ListBills(m.ctx)
	if err != nil {
		return errMsg{err: err}
	}

	return billsLoadedMsg{bills: bills}
}

func (m *BillsModel) editBill() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.bills) {
		return m, nil
	}

	bill := m.bills[m.selectedIndex]
	m.viewMode = BillViewForm
	m.formModel.editing = true
	m.formModel.editingID = &bill.ID

	m.formModel.nameInput = bill.Name
	m.formModel.descriptionInput = bill.Description
	m.formModel.amountInput = fmt.Sprintf("%.2f", bill.TotalAmount.Amount())
	m.formModel.startDateInput = bill.StartDate.Format("2006-01-02")
	m.formModel.endDateInput = bill.EndDate.Format("2006-01-02")
	m.formModel.dueDateInput = bill.DueDate.Format("2006-01-02")

	return m, nil
}

func (m *BillsModel) startPayment() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.bills) {
		return m, nil
	}

	bill := m.bills[m.selectedIndex]

	// Check if bill can receive payments
	if bill.Status == entity.BillStatusClosed || bill.Status == entity.BillStatusPaid {
		m.err = fmt.Errorf("cannot add payment to %s bill", bill.Status)
		return m, nil
	}

	m.viewMode = BillViewPayment
	m.paymentModel = &BillPaymentFormModel{
		billID: bill.ID,
		bill:   bill,
	}

	return m, nil
}

func (m *BillsModel) closeBill() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.bills) {
		return m, nil
	}

	bill := m.bills[m.selectedIndex]

	m.loading = true
	return m, func() tea.Msg {
		err := m.billUseCase.CloseBill(m.ctx, bill.ID)
		if err != nil {
			return errMsg{err: err}
		}
		return billActionMsg{}
	}
}

func (m *BillsModel) deleteBill() tea.Msg {
	if m.selectedIndex >= len(m.bills) {
		return errMsg{err: fmt.Errorf("no bill selected")}
	}

	// TODO: Implement DeleteBill in use case
	// For now, return an error
	return errMsg{err: fmt.Errorf("bill deletion not implemented in backend")}
}

func (m *BillsModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate required fields
	if strings.TrimSpace(m.formModel.nameInput) == "" {
		m.err = fmt.Errorf("bill name is required")
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.formModel.amountInput, 64)
	if err != nil {
		m.err = fmt.Errorf("invalid amount")
		return m, nil
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", m.formModel.startDateInput)
	if err != nil {
		m.err = fmt.Errorf("invalid start date format (use YYYY-MM-DD)")
		return m, nil
	}

	endDate, err := time.Parse("2006-01-02", m.formModel.endDateInput)
	if err != nil {
		m.err = fmt.Errorf("invalid end date format (use YYYY-MM-DD)")
		return m, nil
	}

	dueDate, err := time.Parse("2006-01-02", m.formModel.dueDateInput)
	if err != nil {
		m.err = fmt.Errorf("invalid due date format (use YYYY-MM-DD)")
		return m, nil
	}

	m.loading = true

	if m.formModel.editing && m.formModel.editingID != nil {
		return m, func() tea.Msg { return m.updateBill(*m.formModel.editingID, amount, startDate, endDate, dueDate) }
	}

	return m, func() tea.Msg { return m.createBill(amount, startDate, endDate, dueDate) }
}

func (m *BillsModel) createBill(amount float64, startDate, endDate, dueDate time.Time) tea.Msg {
	_, err := m.billUseCase.CreateBill(
		m.ctx,
		m.formModel.nameInput,
		m.formModel.descriptionInput,
		startDate,
		endDate,
		dueDate,
		amount,
		"BRL",
	)
	if err != nil {
		return errMsg{err: err}
	}

	return billActionMsg{}
}

func (m *BillsModel) updateBill(id uuid.UUID, amount float64, startDate, endDate, dueDate time.Time) tea.Msg {
	// TODO: Implement UpdateBill in use case
	// For now, return an error
	return errMsg{err: fmt.Errorf("bill update not implemented in backend")}
}

func (m *BillsModel) submitPayment() (tea.Model, tea.Cmd) {
	if m.paymentModel == nil {
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.paymentModel.amountInput, 64)
	if err != nil || amount <= 0 {
		m.err = fmt.Errorf("invalid payment amount")
		return m, nil
	}

	m.loading = true
	return m, func() tea.Msg {
		err := m.billUseCase.AddPayment(m.ctx, m.paymentModel.billID, amount, "BRL")
		if err != nil {
			return errMsg{err: err}
		}
		return billActionMsg{}
	}
}

func (m *BillsModel) resetForm() {
	m.formModel.nameInput = ""
	m.formModel.descriptionInput = ""
	m.formModel.amountInput = ""
	m.formModel.startDateInput = ""
	m.formModel.endDateInput = ""
	m.formModel.dueDateInput = ""
	m.formModel.focusedField = 0
}

func (m *BillsModel) IsInFormMode() bool {
	return m.viewMode == BillViewForm || m.viewMode == BillViewPayment || m.viewMode == BillViewConfirm
}
