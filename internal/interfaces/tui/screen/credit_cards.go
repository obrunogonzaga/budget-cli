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

type CreditCardsModel struct {
	ctx               context.Context
	creditCardUseCase *usecase.CreditCardUseCase
	accountUseCase    *usecase.AccountUseCase
	
	// Data
	creditCards       []*entity.CreditCard
	accounts          []*entity.Account
	
	// View state
	selectedIndex     int
	viewMode          CreditCardViewMode
	
	// Loading and errors
	loading           bool
	err               error
	
	// Form state
	formModel         *CreditCardFormModel
	
	// Payment state
	paymentModel      *PaymentFormModel
	
	// Confirmation state
	showConfirmDelete bool
	confirmMessage    string
	
	// Window dimensions
	width             int
	height            int
}

type CreditCardViewMode int

const (
	CreditCardViewList CreditCardViewMode = iota
	CreditCardViewForm
	CreditCardViewDetails
	CreditCardViewPayment
	CreditCardViewConfirm
)

type CreditCardFormModel struct {
	// Form fields
	name              string
	lastFourDigits    string
	creditLimit       string
	dueDay            string
	
	// Selection states
	selectedAccount   int
	
	// Input fields
	nameInput         string
	lastFourInput     string
	limitInput        string
	dueDayInput       string
	
	// Navigation
	focusedField      int
	
	// Edit state
	editing           bool
	editingID         *uuid.UUID
}

type PaymentFormModel struct {
	cardID            uuid.UUID
	card              *entity.CreditCard
	
	// Payment amount
	amountInput       string
	
	// Navigation
	focusedField      int
}

func NewCreditCardsModel(ctx context.Context, creditCardUC *usecase.CreditCardUseCase, accountUC *usecase.AccountUseCase) tea.Model {
	return &CreditCardsModel{
		ctx:               ctx,
		creditCardUseCase: creditCardUC,
		accountUseCase:    accountUC,
		viewMode:          CreditCardViewList,
		loading:           true,
		formModel: &CreditCardFormModel{
			dueDayInput: "1",
		},
		paymentModel: &PaymentFormModel{},
	}
}

func (m *CreditCardsModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadCreditCards,
		m.loadAccounts,
	)
}

func (m *CreditCardsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case cardsLoadedMsg:
		m.loading = false
		m.creditCards = msg.creditCards
		if len(m.creditCards) > 0 && m.selectedIndex >= len(m.creditCards) {
			m.selectedIndex = len(m.creditCards) - 1
		}
		return m, nil
		
	case accountsLoadedMsg:
		m.accounts = msg.accounts
		return m, nil
		
	case creditCardActionMsg:
		m.loading = false
		m.viewMode = CreditCardViewList
		m.resetForm()
		m.resetPaymentForm()
		return m, m.loadCreditCards
		
	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
		
	case tea.KeyMsg:
		switch m.viewMode {
		case CreditCardViewList:
			return m.handleListKeys(msg)
		case CreditCardViewForm:
			return m.handleFormKeys(msg)
		case CreditCardViewDetails:
			return m.handleDetailsKeys(msg)
		case CreditCardViewPayment:
			return m.handlePaymentKeys(msg)
		case CreditCardViewConfirm:
			return m.handleConfirmKeys(msg)
		}
	}
	
	return m, nil
}

func (m *CreditCardsModel) View() string {
	if m.loading {
		return style.InfoStyle.Render("Loading credit cards...")
	}
	
	if m.err != nil {
		return style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	
	switch m.viewMode {
	case CreditCardViewList:
		return m.renderCreditCardsList()
	case CreditCardViewForm:
		return m.renderCreditCardForm()
	case CreditCardViewDetails:
		return m.renderCreditCardDetails()
	case CreditCardViewPayment:
		return m.renderPaymentForm()
	case CreditCardViewConfirm:
		return m.renderConfirmDialog()
	}
	
	return ""
}

// Helper functions for loading data
func (m *CreditCardsModel) loadCreditCards() tea.Msg {
	cards, err := m.creditCardUseCase.ListCreditCards(m.ctx)
	if err != nil {
		return errMsg{err: err}
	}
	
	return cardsLoadedMsg{creditCards: cards}
}

func (m *CreditCardsModel) loadAccounts() tea.Msg {
	accounts, err := m.accountUseCase.ListAccounts(m.ctx)
	if err != nil {
		// Don't fail the whole screen if accounts can't be loaded
		return accountsLoadedMsg{accounts: []*entity.Account{}}
	}
	return accountsLoadedMsg{accounts: accounts}
}

// Helper to reset form
func (m *CreditCardsModel) resetForm() {
	m.formModel = &CreditCardFormModel{
		dueDayInput:   "1",
		focusedField:  0,
		selectedAccount: 0,
	}
}

// Helper to reset payment form
func (m *CreditCardsModel) resetPaymentForm() {
	m.paymentModel = &PaymentFormModel{
		amountInput:  "",
		focusedField: 0,
	}
}

// Get account name by ID
func (m *CreditCardsModel) getAccountName(accountID uuid.UUID) string {
	for _, acc := range m.accounts {
		if acc.ID == accountID {
			return acc.Name
		}
	}
	return "Unknown Account"
}

// Calculate next due date
func (m *CreditCardsModel) getNextDueDate(dueDay int) time.Time {
	now := time.Now()
	year, month, _ := now.Date()
	
	// Create date for this month
	dueDate := time.Date(year, month, dueDay, 0, 0, 0, 0, now.Location())
	
	// If the due date has passed this month, move to next month
	if dueDate.Before(now) {
		if month == time.December {
			dueDate = time.Date(year+1, time.January, dueDay, 0, 0, 0, 0, now.Location())
		} else {
			dueDate = time.Date(year, month+1, dueDay, 0, 0, 0, 0, now.Location())
		}
	}
	
	// Handle months with fewer days than the due day
	for dueDate.Day() != dueDay {
		dueDate = dueDate.AddDate(0, 0, -1)
	}
	
	return dueDate
}

// Get utilization color based on percentage
func (m *CreditCardsModel) getUtilizationColor(percentage float64) lipgloss.Color {
	if percentage < 30 {
		return style.Success
	} else if percentage < 70 {
		return style.Warning
	} else {
		return style.Danger
	}
}

// Render progress bar for utilization
func (m *CreditCardsModel) renderProgressBar(percentage float64, width int) string {
	if width <= 0 {
		width = 10
	}
	
	filled := int(percentage * float64(width) / 100)
	if filled > width {
		filled = width
	}
	
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	
	color := m.getUtilizationColor(percentage)
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

// FormModeChecker interface implementation
func (m *CreditCardsModel) IsInFormMode() bool {
	return m.viewMode != CreditCardViewList
}

// Message types
type cardsLoadedMsg struct {
	creditCards []*entity.CreditCard
}

type creditCardActionMsg struct{}

// Key handler for list view
func (m *CreditCardsModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.creditCards)-1 {
			m.selectedIndex++
		}
	case "enter":
		if len(m.creditCards) > 0 {
			m.viewMode = CreditCardViewDetails
		}
	case "n":
		m.viewMode = CreditCardViewForm
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
	case "e":
		if len(m.creditCards) > 0 {
			return m.editCreditCard()
		}
	case "d":
		if len(m.creditCards) > 0 {
			m.viewMode = CreditCardViewConfirm
			m.showConfirmDelete = true
		}
	case "p":
		if len(m.creditCards) > 0 && m.selectedIndex < len(m.creditCards) {
			card := m.creditCards[m.selectedIndex]
			m.paymentModel.cardID = card.ID
			m.paymentModel.card = card
			m.viewMode = CreditCardViewPayment
		}
	case "r":
		m.loading = true
		return m, tea.Batch(
			m.loadCreditCards,
			m.loadAccounts,
		)
	case "b":
		return m, func() tea.Msg { return BackToDashboardMsg{} }
	}
	
	return m, nil
}

func (m *CreditCardsModel) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalFields := 6 // name, last4, limit, account, dueday, buttons
	
	switch msg.String() {
	case "esc":
		m.viewMode = CreditCardViewList
		m.resetForm()
	case "tab", "down":
		m.formModel.focusedField = (m.formModel.focusedField + 1) % totalFields
	case "shift+tab", "up":
		m.formModel.focusedField = (m.formModel.focusedField - 1 + totalFields) % totalFields
	case "enter":
		if m.formModel.focusedField == 4 {
			// Submit button
			return m.submitForm()
		} else if m.formModel.focusedField == 5 {
			// Cancel button
			m.viewMode = CreditCardViewList
			m.resetForm()
		}
	default:
		return m.handleFormInput(msg)
	}
	
	return m, nil
}

func (m *CreditCardsModel) handleDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "b", "enter":
		m.viewMode = CreditCardViewList
	case "e":
		return m.editCreditCard()
	case "d":
		m.viewMode = CreditCardViewConfirm
		m.showConfirmDelete = true
	case "p":
		if m.selectedIndex < len(m.creditCards) {
			card := m.creditCards[m.selectedIndex]
			m.paymentModel.cardID = card.ID
			m.paymentModel.card = card
			m.viewMode = CreditCardViewPayment
		}
	}
	
	return m, nil
}

func (m *CreditCardsModel) handlePaymentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = CreditCardViewDetails
		m.resetPaymentForm()
	case "tab", "down":
		m.paymentModel.focusedField = (m.paymentModel.focusedField + 1) % 2
	case "shift+tab", "up":
		m.paymentModel.focusedField = (m.paymentModel.focusedField - 1 + 2) % 2
	case "enter":
		if m.paymentModel.focusedField == 0 {
			// Submit payment
			return m.submitPayment()
		} else {
			// Cancel button
			m.viewMode = CreditCardViewDetails
			m.resetPaymentForm()
		}
	case "backspace":
		if m.paymentModel.focusedField == 0 && len(m.paymentModel.amountInput) > 0 {
			m.paymentModel.amountInput = m.paymentModel.amountInput[:len(m.paymentModel.amountInput)-1]
		}
	default:
		if m.paymentModel.focusedField == 0 && len(msg.String()) == 1 {
			if msg.String() >= "0" && msg.String() <= "9" || msg.String() == "." {
				m.paymentModel.amountInput += msg.String()
			}
		}
	}
	
	return m, nil
}

func (m *CreditCardsModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.viewMode = CreditCardViewList
		m.showConfirmDelete = false
		m.loading = true
		return m, m.deleteCreditCard
	case "n", "esc":
		m.viewMode = CreditCardViewList
		m.showConfirmDelete = false
	}
	
	return m, nil
}

// Helper method to edit a credit card
func (m *CreditCardsModel) editCreditCard() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.creditCards) {
		return m, nil
	}
	
	card := m.creditCards[m.selectedIndex]
	m.viewMode = CreditCardViewForm
	m.formModel.editing = true
	m.formModel.editingID = &card.ID
	
	// Pre-fill form with card data
	m.formModel.nameInput = card.Name
	m.formModel.lastFourInput = card.LastFourDigits
	m.formModel.limitInput = fmt.Sprintf("%.2f", card.CreditLimit.Amount())
	m.formModel.dueDayInput = strconv.Itoa(card.DueDay)
	
	// Find and set the account
	for i, acc := range m.accounts {
		if acc.ID == card.AccountID {
			m.formModel.selectedAccount = i
			break
		}
	}
	
	return m, nil
}

// View rendering for credit cards list
func (m *CreditCardsModel) renderCreditCardsList() string {
	var sections []string
	
	title := style.TitleStyle.Render("💳 Credit Cards Management")
	sections = append(sections, title)
	
	// Summary section
	summary := m.renderSummary()
	sections = append(sections, summary)
	
	if len(m.creditCards) == 0 {
		empty := style.InfoStyle.Render("No credit cards found. Press 'n' to add your first credit card.")
		sections = append(sections, empty)
	} else {
		// Credit cards table
		table := m.renderCreditCardsTable()
		sections = append(sections, table)
		
		// Selected card details
		if m.selectedIndex < len(m.creditCards) {
			details := m.renderSelectedCardInfo()
			sections = append(sections, details)
		}
	}
	
	help := m.renderListHelp()
	sections = append(sections, help)
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render summary section
func (m *CreditCardsModel) renderSummary() string {
	var totalLimit, totalBalance, totalAvailable float64
	cardCount := len(m.creditCards)
	
	for _, card := range m.creditCards {
		totalLimit += card.CreditLimit.Amount()
		totalBalance += card.CurrentBalance.Amount()
		available, _ := card.GetAvailableCredit()
		totalAvailable += available.Amount()
	}
	
	avgUtilization := 0.0
	if totalLimit > 0 {
		avgUtilization = (totalBalance / totalLimit) * 100
	}
	
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(0, 1).
		MarginTop(1)
	
	cards := fmt.Sprintf("Cards: %d", cardCount)
	limit := fmt.Sprintf("Total Limit: R$ %.2f", totalLimit)
	balance := fmt.Sprintf("Total Balance: R$ %.2f", totalBalance)
	available := style.SuccessStyle.Render(fmt.Sprintf("Available: R$ %.2f", totalAvailable))
	
	utilColor := m.getUtilizationColor(avgUtilization)
	utilization := lipgloss.NewStyle().Foreground(utilColor).Render(fmt.Sprintf("Avg Utilization: %.1f%%", avgUtilization))
	
	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		cards,
		" | ",
		limit,
		" | ",
		balance,
		" | ",
		available,
		" | ",
		utilization,
	)
	
	return summaryStyle.Render(content)
}

// Render credit cards table
func (m *CreditCardsModel) renderCreditCardsTable() string {
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)
	
	headers := []string{"Card Name", "Last 4", "Balance", "Limit", "Available", "Utilization"}
	headerRow := style.TableHeaderStyle.Render(
		fmt.Sprintf("%-20s %-8s %-12s %-12s %-12s %-25s",
			headers[0], headers[1], headers[2], headers[3], headers[4], headers[5]),
	)
	
	var rows []string
	rows = append(rows, headerRow)
	
	for i, card := range m.creditCards {
		name := truncateString(card.Name, 20)
		lastFour := card.LastFourDigits
		balance := fmt.Sprintf("R$ %.2f", card.CurrentBalance.Amount())
		limit := fmt.Sprintf("R$ %.2f", card.CreditLimit.Amount())
		
		available, _ := card.GetAvailableCredit()
		availableStr := fmt.Sprintf("R$ %.2f", available.Amount())
		
		utilization := card.GetUtilizationPercentage()
		utilizationBar := m.renderProgressBar(utilization, 10)
		utilizationStr := fmt.Sprintf("%s %.1f%%", utilizationBar, utilization)
		
		row := fmt.Sprintf("%-20s %-8s %-12s %-12s %-12s %-25s",
			name, lastFour, balance, limit, availableStr, utilizationStr)
		
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

// Render selected card info
func (m *CreditCardsModel) renderSelectedCardInfo() string {
	if m.selectedIndex >= len(m.creditCards) {
		return ""
	}
	
	card := m.creditCards[m.selectedIndex]
	
	detailsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)
	
	accountName := m.getAccountName(card.AccountID)
	nextDue := m.getNextDueDate(card.DueDay)
	daysUntilDue := int(time.Until(nextDue).Hours() / 24)
	
	details := []string{
		fmt.Sprintf("Linked Account: %s", accountName),
		fmt.Sprintf("Due Day: %d of each month", card.DueDay),
		fmt.Sprintf("Next Due Date: %s (%d days)", nextDue.Format("Jan 2, 2006"), daysUntilDue),
		fmt.Sprintf("Created: %s", card.CreatedAt.Format("2006-01-02")),
	}
	
	content := strings.Join(details, "\n")
	return detailsStyle.Render(content)
}

// Render help text for list view
func (m *CreditCardsModel) renderListHelp() string {
	help := "[↑/↓] Navigate • [Enter] Details • [n] New • [e] Edit • [d] Delete • [p] Payment • [r] Refresh • [b] Back"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}


// Handle form input based on focused field
func (m *CreditCardsModel) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case 1: // Last 4 digits
		switch msg.String() {
		case "backspace":
			if len(m.formModel.lastFourInput) > 0 {
				m.formModel.lastFourInput = m.formModel.lastFourInput[:len(m.formModel.lastFourInput)-1]
			}
		default:
			if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" && len(m.formModel.lastFourInput) < 4 {
				m.formModel.lastFourInput += msg.String()
			}
		}
	case 2: // Credit limit
		switch msg.String() {
		case "backspace":
			if len(m.formModel.limitInput) > 0 {
				m.formModel.limitInput = m.formModel.limitInput[:len(m.formModel.limitInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == ".") {
				m.formModel.limitInput += msg.String()
			}
		}
	case 3: // Account selection
		switch msg.String() {
		case "left":
			if m.formModel.selectedAccount > 0 {
				m.formModel.selectedAccount--
			}
		case "right":
			if m.formModel.selectedAccount < len(m.accounts)-1 {
				m.formModel.selectedAccount++
			}
		}
	case 4: // Due day
		switch msg.String() {
		case "backspace":
			if len(m.formModel.dueDayInput) > 0 {
				m.formModel.dueDayInput = m.formModel.dueDayInput[:len(m.formModel.dueDayInput)-1]
			}
		default:
			if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" && len(m.formModel.dueDayInput) < 2 {
				m.formModel.dueDayInput += msg.String()
			}
		}
	}
	
	return m, nil
}

// Submit credit card form
func (m *CreditCardsModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate form
	if strings.TrimSpace(m.formModel.nameInput) == "" {
		m.err = fmt.Errorf("card name is required")
		return m, nil
	}
	
	if len(m.formModel.lastFourInput) != 4 {
		m.err = fmt.Errorf("last 4 digits must be exactly 4 numbers")
		return m, nil
	}
	
	limit, err := strconv.ParseFloat(m.formModel.limitInput, 64)
	if err != nil || limit <= 0 {
		m.err = fmt.Errorf("invalid credit limit")
		return m, nil
	}
	
	dueDay, err := strconv.Atoi(m.formModel.dueDayInput)
	if err != nil || dueDay < 1 || dueDay > 31 {
		m.err = fmt.Errorf("due day must be between 1 and 31")
		return m, nil
	}
	
	if len(m.accounts) == 0 {
		m.err = fmt.Errorf("no accounts available to link")
		return m, nil
	}
	
	accountID := m.accounts[m.formModel.selectedAccount].ID
	
	m.loading = true
	
	if m.formModel.editing && m.formModel.editingID != nil {
		// TODO: Implement update credit card
		return m, func() tea.Msg {
			return errMsg{err: fmt.Errorf("update credit card not implemented yet")}
		}
	}
	
	// Create credit card
	return m, func() tea.Msg {
		_, err := m.creditCardUseCase.CreateCreditCard(
			m.ctx,
			accountID,
			m.formModel.nameInput,
			m.formModel.lastFourInput,
			limit,
			"BRL",
			dueDay,
		)
		if err != nil {
			return errMsg{err: err}
		}
		
		return creditCardActionMsg{}
	}
}

func (m *CreditCardsModel) renderCreditCardForm() string {
	var sections []string
	
	title := "📝 New Credit Card"
	if m.formModel.editing {
		title = "✏️ Edit Credit Card"
	}
	sections = append(sections, style.TitleStyle.Render(title))
	
	if m.err != nil {
		errorMsg := style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		sections = append(sections, errorMsg)
	}
	
	form := m.renderForm()
	sections = append(sections, form)
	
	help := m.renderFormHelp()
	sections = append(sections, help)
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render the credit card form
func (m *CreditCardsModel) renderForm() string {
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(2, 4).
		MarginTop(1)
	
	var fields []string
	
	// Name field
	fields = append(fields, m.renderFormField("Card Name:", m.formModel.nameInput, 0))
	
	// Last 4 digits field
	fields = append(fields, m.renderFormField("Last 4 Digits:", m.formModel.lastFourInput, 1))
	
	// Credit limit field
	fields = append(fields, m.renderFormField("Credit Limit (R$):", m.formModel.limitInput, 2))
	
	// Account selector
	fields = append(fields, m.renderAccountSelector())
	
	// Due day field
	fields = append(fields, m.renderFormField("Due Day (1-31):", m.formModel.dueDayInput, 4))
	
	// Buttons
	buttons := m.renderFormButtons()
	fields = append(fields, buttons)
	
	content := strings.Join(fields, "\n\n")
	return formStyle.Render(content)
}

// Render a form field
func (m *CreditCardsModel) renderFormField(label, value string, fieldIndex int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)
	
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

// Render account selector
func (m *CreditCardsModel) renderAccountSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)
	
	if len(m.accounts) == 0 {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Linked Account:"),
			style.WarningStyle.Render("No accounts available"),
		)
	}
	
	account := m.accounts[m.formModel.selectedAccount]
	display := fmt.Sprintf("%s (R$ %.2f)", account.Name, account.Balance.Amount())
	
	var selector string
	if m.formModel.focusedField == 3 {
		selector = style.FocusedInputStyle.Width(30).Render("< " + display + " >")
		selector = selector + " ◄"
	} else {
		selector = style.InputStyle.Width(30).Render(display)
	}
	
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Linked Account:"),
		selector,
	)
}

// Render form buttons
func (m *CreditCardsModel) renderFormButtons() string {
	submitText := "Create Card"
	if m.formModel.editing {
		submitText = "Update Card"
	}
	
	var submitStyle, cancelStyle lipgloss.Style
	
	// Submit button styling
	if m.formModel.focusedField == 4 {
		submitStyle = style.ButtonStyle.Background(style.Success)
	} else {
		submitStyle = style.SecondaryButtonStyle
	}
	
	// Cancel button styling
	if m.formModel.focusedField == 5 {
		cancelStyle = style.ButtonStyle.Background(style.Danger)
	} else {
		cancelStyle = style.SecondaryButtonStyle
	}
	
	submitBtn := submitStyle.Render(submitText)
	cancelBtn := cancelStyle.Render("Cancel")
	
	// Add focus indicators
	if m.formModel.focusedField == 4 {
		submitBtn = submitBtn + " ◄"
	} else if m.formModel.focusedField == 5 {
		cancelBtn = cancelBtn + " ◄"
	}
	
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		submitBtn,
		lipgloss.NewStyle().MarginLeft(2).Render(cancelBtn),
	)
}

// Render form help
func (m *CreditCardsModel) renderFormHelp() string {
	help := "[Tab] Next Field • [Shift+Tab] Previous • [←/→] Select Account • [Enter] Confirm • [Esc] Cancel"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

func (m *CreditCardsModel) renderCreditCardDetails() string {
	if m.selectedIndex >= len(m.creditCards) {
		return style.ErrorStyle.Render("No credit card selected")
	}
	
	card := m.creditCards[m.selectedIndex]
	
	var sections []string
	
	title := style.TitleStyle.Render("💳 Credit Card Details")
	sections = append(sections, title)
	
	// Main card info
	mainInfo := m.renderCardMainInfo(card)
	sections = append(sections, mainInfo)
	
	// Utilization visualization
	utilization := m.renderUtilizationVisualization(card)
	sections = append(sections, utilization)
	
	// Account and due date info
	additionalInfo := m.renderCardAdditionalInfo(card)
	sections = append(sections, additionalInfo)
	
	// Actions
	actions := m.renderDetailsActions()
	sections = append(sections, actions)
	
	help := "[Esc/Enter] Back • [p] Make Payment • [e] Edit • [d] Delete"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render main card information
func (m *CreditCardsModel) renderCardMainInfo(card *entity.CreditCard) string {
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)
	
	var info []string
	
	info = append(info, style.HeaderStyle.Render("Card Information"))
	info = append(info, fmt.Sprintf("Name: %s", card.Name))
	info = append(info, fmt.Sprintf("Last 4 Digits: •••• %s", card.LastFourDigits))
	info = append(info, fmt.Sprintf("Credit Limit: R$ %.2f", card.CreditLimit.Amount()))
	info = append(info, fmt.Sprintf("Current Balance: R$ %.2f", card.CurrentBalance.Amount()))
	
	available, _ := card.GetAvailableCredit()
	info = append(info, fmt.Sprintf("Available Credit: %s", 
		style.SuccessStyle.Render(fmt.Sprintf("R$ %.2f", available.Amount()))))
	
	content := strings.Join(info, "\n")
	return infoStyle.Render(content)
}

// Render utilization visualization
func (m *CreditCardsModel) renderUtilizationVisualization(card *entity.CreditCard) string {
	utilizationStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Warning).
		Padding(1, 2).
		MarginTop(1)
	
	var content []string
	
	content = append(content, style.HeaderStyle.Render("Credit Utilization"))
	
	utilization := card.GetUtilizationPercentage()
	
	// Large progress bar
	bar := m.renderProgressBar(utilization, 40)
	content = append(content, bar)
	
	// Percentage and amount
	color := m.getUtilizationColor(utilization)
	percentStr := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(fmt.Sprintf("%.1f%%", utilization))
	
	details := fmt.Sprintf("%s - R$ %.2f of R$ %.2f used", 
		percentStr, 
		card.CurrentBalance.Amount(),
		card.CreditLimit.Amount())
	
	content = append(content, details)
	
	// Utilization status message
	var status string
	if utilization < 30 {
		status = style.SuccessStyle.Render("✓ Excellent - Low utilization")
	} else if utilization < 70 {
		status = style.WarningStyle.Render("! Good - Moderate utilization")
	} else {
		status = style.ErrorStyle.Render("⚠ High utilization - Consider paying down balance")
	}
	content = append(content, "", status)
	
	return utilizationStyle.Render(strings.Join(content, "\n"))
}

// Render additional card information
func (m *CreditCardsModel) renderCardAdditionalInfo(card *entity.CreditCard) string {
	additionalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)
	
	var info []string
	
	info = append(info, style.HeaderStyle.Render("Additional Information"))
	
	// Linked account
	accountName := m.getAccountName(card.AccountID)
	info = append(info, fmt.Sprintf("Linked Account: %s", accountName))
	
	// Due date information
	info = append(info, fmt.Sprintf("Due Day: %d of each month", card.DueDay))
	
	nextDue := m.getNextDueDate(card.DueDay)
	daysUntilDue := int(time.Until(nextDue).Hours() / 24)
	
	var dueStatus string
	if daysUntilDue <= 3 {
		dueStatus = style.ErrorStyle.Render(fmt.Sprintf("%d days", daysUntilDue))
	} else if daysUntilDue <= 7 {
		dueStatus = style.WarningStyle.Render(fmt.Sprintf("%d days", daysUntilDue))
	} else {
		dueStatus = style.InfoStyle.Render(fmt.Sprintf("%d days", daysUntilDue))
	}
	
	info = append(info, fmt.Sprintf("Next Due Date: %s (%s)", 
		nextDue.Format("Monday, Jan 2, 2006"), dueStatus))
	
	// Timestamps
	info = append(info, "")
	info = append(info, fmt.Sprintf("Created: %s", card.CreatedAt.Format("2006-01-02 15:04")))
	info = append(info, fmt.Sprintf("Updated: %s", card.UpdatedAt.Format("2006-01-02 15:04")))
	
	content := strings.Join(info, "\n")
	return additionalStyle.Render(content)
}

// Render details actions
func (m *CreditCardsModel) renderDetailsActions() string {
	actionsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Primary).
		Padding(1, 2).
		MarginTop(1)
	
	title := style.HeaderStyle.Render("Quick Actions")
	
	actions := []string{
		"[p] Make Payment - Pay down your balance",
		"[e] Edit Card - Update card information",
		"[d] Delete Card - Remove this credit card",
	}
	
	content := title + "\n" + strings.Join(actions, "\n")
	return actionsStyle.Render(content)
}

// Submit payment
func (m *CreditCardsModel) submitPayment() (tea.Model, tea.Cmd) {
	amount, err := strconv.ParseFloat(m.paymentModel.amountInput, 64)
	if err != nil || amount <= 0 {
		m.err = fmt.Errorf("invalid payment amount")
		return m, nil
	}
	
	if m.paymentModel.card == nil {
		m.err = fmt.Errorf("no card selected for payment")
		return m, nil
	}
	
	// Check if amount exceeds current balance
	if amount > m.paymentModel.card.CurrentBalance.Amount() {
		amount = m.paymentModel.card.CurrentBalance.Amount()
	}
	
	m.loading = true
	
	return m, func() tea.Msg {
		err := m.creditCardUseCase.MakePayment(
			m.ctx,
			m.paymentModel.cardID,
			amount,
			"BRL",
		)
		if err != nil {
			return errMsg{err: err}
		}
		
		return creditCardActionMsg{}
	}
}

func (m *CreditCardsModel) renderPaymentForm() string {
	if m.paymentModel.card == nil {
		return style.ErrorStyle.Render("No card selected for payment")
	}
	
	var sections []string
	
	title := style.TitleStyle.Render("💰 Make Payment")
	sections = append(sections, title)
	
	// Card info
	cardInfo := m.renderPaymentCardInfo()
	sections = append(sections, cardInfo)
	
	// Payment form
	form := m.renderPaymentFormFields()
	sections = append(sections, form)
	
	if m.err != nil {
		errorMsg := style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		sections = append(sections, errorMsg)
	}
	
	help := "[Tab] Navigate • [Enter] Submit • [Esc] Cancel"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render payment card info
func (m *CreditCardsModel) renderPaymentCardInfo() string {
	card := m.paymentModel.card
	
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)
	
	var info []string
	
	info = append(info, style.HeaderStyle.Render("Card Information"))
	info = append(info, fmt.Sprintf("Card: %s (•••• %s)", card.Name, card.LastFourDigits))
	info = append(info, fmt.Sprintf("Current Balance: %s", 
		style.ErrorStyle.Render(fmt.Sprintf("R$ %.2f", card.CurrentBalance.Amount()))))
	
	available, _ := card.GetAvailableCredit()
	info = append(info, fmt.Sprintf("Available After Payment: R$ %.2f", available.Amount()))
	
	// Get linked account
	var accountBalance float64
	accountName := "Unknown Account"
	for _, acc := range m.accounts {
		if acc.ID == card.AccountID {
			accountName = acc.Name
			accountBalance = acc.Balance.Amount()
			break
		}
	}
	
	info = append(info, "")
	info = append(info, fmt.Sprintf("Payment From: %s", accountName))
	info = append(info, fmt.Sprintf("Account Balance: R$ %.2f", accountBalance))
	
	content := strings.Join(info, "\n")
	return infoStyle.Render(content)
}

// Render payment form fields
func (m *CreditCardsModel) renderPaymentFormFields() string {
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(2, 4).
		MarginTop(1)
	
	var fields []string
	
	// Amount input
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
	
	amountField := lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Payment Amount (R$):"),
		input,
	)
	fields = append(fields, amountField)
	
	// Quick amount suggestions
	if m.paymentModel.card != nil {
		balance := m.paymentModel.card.CurrentBalance.Amount()
		suggestions := []string{
			fmt.Sprintf("Full Balance: R$ %.2f", balance),
			fmt.Sprintf("Minimum: R$ %.2f", balance*0.1),
			fmt.Sprintf("Half: R$ %.2f", balance*0.5),
		}
		suggestionText := style.InfoStyle.Render(strings.Join(suggestions, " | "))
		fields = append(fields, suggestionText)
	}
	
	// Buttons
	fields = append(fields, "")
	
	var submitStyle, cancelStyle lipgloss.Style
	
	if m.paymentModel.focusedField == 0 {
		submitStyle = style.ButtonStyle.Background(style.Success)
		cancelStyle = style.SecondaryButtonStyle
	} else {
		submitStyle = style.SecondaryButtonStyle
		cancelStyle = style.ButtonStyle.Background(style.Danger)
	}
	
	submitBtn := submitStyle.Render("Make Payment")
	cancelBtn := cancelStyle.Render("Cancel")
	
	if m.paymentModel.focusedField == 0 {
		submitBtn = submitBtn + " ◄"
	} else {
		cancelBtn = cancelBtn + " ◄"
	}
	
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		submitBtn,
		lipgloss.NewStyle().MarginLeft(2).Render(cancelBtn),
	)
	fields = append(fields, buttons)
	
	content := strings.Join(fields, "\n")
	return formStyle.Render(content)
}

// Delete credit card
func (m *CreditCardsModel) deleteCreditCard() tea.Msg {
	if m.selectedIndex >= len(m.creditCards) {
		return errMsg{err: fmt.Errorf("no credit card selected")}
	}
	
	_ = m.creditCards[m.selectedIndex]
	
	// TODO: Implement delete in use case
	// For now, we just return an error
	return errMsg{err: fmt.Errorf("delete credit card not implemented yet")}
}

func (m *CreditCardsModel) renderConfirmDialog() string {
	if m.selectedIndex >= len(m.creditCards) {
		return style.ErrorStyle.Render("No credit card selected")
	}
	
	card := m.creditCards[m.selectedIndex]
	
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Danger).
		Padding(2, 4).
		MarginTop(5)
	
	title := style.ErrorStyle.Render("⚠️  Confirm Delete")
	
	available, _ := card.GetAvailableCredit()
	
	var warningMessages []string
	warningMessages = append(warningMessages, 
		fmt.Sprintf("Are you sure you want to delete credit card '%s'?", card.Name))
	warningMessages = append(warningMessages, "")
	warningMessages = append(warningMessages, 
		fmt.Sprintf("Card: •••• %s", card.LastFourDigits))
	warningMessages = append(warningMessages, 
		fmt.Sprintf("Current Balance: R$ %.2f", card.CurrentBalance.Amount()))
	warningMessages = append(warningMessages, 
		fmt.Sprintf("Available Credit: R$ %.2f", available.Amount()))
	
	// Add extra warning if there's a balance
	if card.CurrentBalance.Amount() > 0 {
		warningMessages = append(warningMessages, "")
		warningMessages = append(warningMessages, 
			style.WarningStyle.Render("⚠️  This card has an outstanding balance!"))
		warningMessages = append(warningMessages, 
			style.WarningStyle.Render("   Please pay off the balance before deleting."))
	}
	
	warningMessages = append(warningMessages, "")
	warningMessages = append(warningMessages, 
		style.WarningStyle.Render("This action cannot be undone!"))
	
	message := strings.Join(warningMessages, "\n")
	help := "[y] Yes, Delete • [n] Cancel"
	
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		help,
	)
	
	return dialogStyle.Render(content)
}