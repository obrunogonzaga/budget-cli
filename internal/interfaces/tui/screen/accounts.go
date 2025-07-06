package screen

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"financli/internal/application/usecase"
	"financli/internal/domain/entity"
	"financli/internal/interfaces/tui/style"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type AccountsModel struct {
	ctx            context.Context
	accountUseCase *usecase.AccountUseCase
	
	accounts       []*entity.Account
	selectedIndex  int
	viewMode       AccountViewMode
	
	loading        bool
	err            error
	
	// Form state
	formModel      *AccountFormModel
	showConfirmDelete bool
	
	width          int
	height         int
}

type AccountViewMode int

const (
	AccountViewList AccountViewMode = iota
	AccountViewForm
	AccountViewConfirm
)

type AccountFormModel struct {
	name        string
	accountType entity.AccountType
	balance     string
	description string
	
	focusedField int
	editing      bool
	editingID    *uuid.UUID
	
	nameInput        string
	balanceInput     string
	descriptionInput string
	typeOptions      []string
	selectedType     int
}

func NewAccountsModel(ctx context.Context, accountUC *usecase.AccountUseCase) tea.Model {
	return &AccountsModel{
		ctx:            ctx,
		accountUseCase: accountUC,
		viewMode:       AccountViewList,
		loading:        true,
		formModel: &AccountFormModel{
			typeOptions: []string{"Checking", "Savings", "Investment"},
		},
	}
}

func (m *AccountsModel) Init() tea.Cmd {
	return m.loadAccounts
}

func (m *AccountsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case accountsLoadedMsg:
		m.loading = false
		m.accounts = msg.accounts
		if len(m.accounts) > 0 && m.selectedIndex >= len(m.accounts) {
			m.selectedIndex = len(m.accounts) - 1
		}
		return m, nil
		
	case accountActionMsg:
		m.loading = false
		m.viewMode = AccountViewList
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
		return m, m.loadAccounts
		
	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
		
	case tea.KeyMsg:
		switch m.viewMode {
		case AccountViewList:
			return m.handleListKeys(msg)
		case AccountViewForm:
			return m.handleFormKeys(msg)
		case AccountViewConfirm:
			return m.handleConfirmKeys(msg)
		}
	}
	
	return m, nil
}

func (m *AccountsModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.accounts)-1 {
			m.selectedIndex++
		}
	case "enter":
		if len(m.accounts) > 0 {
			return m.showAccountDetails()
		}
	case "n":
		m.viewMode = AccountViewForm
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
	case "e":
		if len(m.accounts) > 0 {
			return m.editAccount()
		}
	case "d":
		if len(m.accounts) > 0 {
			m.viewMode = AccountViewConfirm
			m.showConfirmDelete = true
		}
	case "r":
		m.loading = true
		return m, m.loadAccounts
	case "b":
		// Go back to dashboard - this will be handled by returning a special message
		return m, func() tea.Msg { return BackToDashboardMsg{} }
	}
	
	return m, nil
}

func (m *AccountsModel) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = AccountViewList
		m.resetForm()
	case "tab", "down":
		m.formModel.focusedField = (m.formModel.focusedField + 1) % 6
	case "shift+tab", "up":
		m.formModel.focusedField = (m.formModel.focusedField - 1 + 6) % 6
	case "enter":
		if m.formModel.focusedField == 4 {
			return m.submitForm()
		} else if m.formModel.focusedField == 5 {
			// Cancel button
			m.viewMode = AccountViewList
			m.resetForm()
		}
	default:
		return m.handleFormInput(msg)
	}
	
	return m, nil
}

func (m *AccountsModel) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only handle input for fields 0-3 (name, type, balance, description)
	// Fields 4-5 are buttons
	if m.formModel.focusedField > 3 {
		return m, nil
	}
	
	switch m.formModel.focusedField {
	case 0:
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
	case 1:
		switch msg.String() {
		case "left":
			if m.formModel.selectedType > 0 {
				m.formModel.selectedType--
			}
		case "right":
			if m.formModel.selectedType < len(m.formModel.typeOptions)-1 {
				m.formModel.selectedType++
			}
		}
	case 2:
		switch msg.String() {
		case "backspace":
			if len(m.formModel.balanceInput) > 0 {
				m.formModel.balanceInput = m.formModel.balanceInput[:len(m.formModel.balanceInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == ".") {
				m.formModel.balanceInput += msg.String()
			}
		}
	case 3:
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
	}
	
	return m, nil
}

func (m *AccountsModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.viewMode = AccountViewList
		m.showConfirmDelete = false
		m.loading = true
		return m, m.deleteAccount
	case "n", "esc":
		m.viewMode = AccountViewList
		m.showConfirmDelete = false
	}
	
	return m, nil
}

func (m *AccountsModel) View() string {
	if m.loading {
		return style.InfoStyle.Render("Loading accounts...")
	}
	
	if m.err != nil {
		return style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	
	switch m.viewMode {
	case AccountViewList:
		return m.renderAccountsList()
	case AccountViewForm:
		return m.renderAccountForm()
	case AccountViewConfirm:
		return m.renderConfirmDialog()
	}
	
	return ""
}

func (m *AccountsModel) renderAccountsList() string {
	var sections []string
	
	title := style.TitleStyle.Render("📊 Accounts Management")
	sections = append(sections, title)
	
	if len(m.accounts) == 0 {
		empty := style.InfoStyle.Render("No accounts found. Press 'n' to create your first account.")
		sections = append(sections, empty)
	} else {
		table := m.renderAccountsTable()
		sections = append(sections, table)
		
		details := m.renderAccountDetails()
		sections = append(sections, details)
	}
	
	help := m.renderListHelp()
	sections = append(sections, help)
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *AccountsModel) renderAccountsTable() string {
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)
	
	headers := []string{"Type", "Account Name", "Balance", "Description"}
	headerRow := style.TableHeaderStyle.Render(
		fmt.Sprintf("%-8s %-20s %-15s %s", headers[0], headers[1], headers[2], headers[3]),
	)
	
	var rows []string
	rows = append(rows, headerRow)
	
	for i, account := range m.accounts {
		icon := m.getAccountIcon(account.Type)
		name := truncateString(account.Name, 20)
		balance := account.Balance.String()
		description := truncateString(account.Description, 30)
		
		row := fmt.Sprintf("%-8s %-20s %-15s %s", 
			icon, name, balance, description)
		
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

func (m *AccountsModel) renderAccountDetails() string {
	if len(m.accounts) == 0 || m.selectedIndex >= len(m.accounts) {
		return ""
	}
	
	account := m.accounts[m.selectedIndex]
	
	detailsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)
	
	details := []string{
		fmt.Sprintf("Account ID: %s", account.ID.String()),
		fmt.Sprintf("Type: %s", m.getAccountTypeName(account.Type)),
		fmt.Sprintf("Created: %s", account.CreatedAt.Format("2006-01-02 15:04")),
		fmt.Sprintf("Updated: %s", account.UpdatedAt.Format("2006-01-02 15:04")),
	}
	
	content := strings.Join(details, "\n")
	return detailsStyle.Render(content)
}

func (m *AccountsModel) renderListHelp() string {
	help := "[↑/↓] Navigate • [Enter] View • [n] New • [e] Edit • [d] Delete • [r] Refresh • [b] Back"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

func (m *AccountsModel) getAccountIcon(accountType entity.AccountType) string {
	switch accountType {
	case entity.AccountTypeChecking:
		return "💳"
	case entity.AccountTypeSavings:
		return "🏦"
	case entity.AccountTypeInvestment:
		return "📈"
	default:
		return "💰"
	}
}

func (m *AccountsModel) getAccountTypeName(accountType entity.AccountType) string {
	switch accountType {
	case entity.AccountTypeChecking:
		return "Checking"
	case entity.AccountTypeSavings:
		return "Savings"
	case entity.AccountTypeInvestment:
		return "Investment"
	default:
		return "Unknown"
	}
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func (m *AccountsModel) showAccountDetails() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *AccountsModel) editAccount() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.accounts) {
		return m, nil
	}
	
	account := m.accounts[m.selectedIndex]
	m.viewMode = AccountViewForm
	m.formModel.editing = true
	m.formModel.editingID = &account.ID
	
	m.formModel.nameInput = account.Name
	m.formModel.balanceInput = fmt.Sprintf("%.2f", account.Balance.Amount())
	m.formModel.descriptionInput = account.Description
	
	switch account.Type {
	case entity.AccountTypeChecking:
		m.formModel.selectedType = 0
	case entity.AccountTypeSavings:
		m.formModel.selectedType = 1
	case entity.AccountTypeInvestment:
		m.formModel.selectedType = 2
	}
	
	return m, nil
}

func (m *AccountsModel) resetForm() {
	m.formModel.nameInput = ""
	m.formModel.balanceInput = ""
	m.formModel.descriptionInput = ""
	m.formModel.selectedType = 0
	m.formModel.focusedField = 0
}

func (m *AccountsModel) loadAccounts() tea.Msg {
	accounts, err := m.accountUseCase.ListAccounts(m.ctx)
	if err != nil {
		return errMsg{err: err}
	}
	
	return accountsLoadedMsg{accounts: accounts}
}

func (m *AccountsModel) submitForm() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.formModel.nameInput) == "" {
		m.err = fmt.Errorf("account name is required")
		return m, nil
	}
	
	balance, err := strconv.ParseFloat(m.formModel.balanceInput, 64)
	if err != nil {
		m.err = fmt.Errorf("invalid balance amount")
		return m, nil
	}
	
	accountType := m.getAccountTypeFromSelection()
	
	m.loading = true
	
	if m.formModel.editing && m.formModel.editingID != nil {
		return m, func() tea.Msg { return m.updateAccount(accountType, balance) }
	}
	
	return m, func() tea.Msg { return m.createAccount(accountType, balance) }
}

func (m *AccountsModel) getAccountTypeFromSelection() entity.AccountType {
	switch m.formModel.selectedType {
	case 0:
		return entity.AccountTypeChecking
	case 1:
		return entity.AccountTypeSavings
	case 2:
		return entity.AccountTypeInvestment
	default:
		return entity.AccountTypeChecking
	}
}

func (m *AccountsModel) createAccount(accountType entity.AccountType, balance float64) tea.Msg {
	_, err := m.accountUseCase.CreateAccount(
		m.ctx,
		m.formModel.nameInput,
		accountType,
		balance,
		"BRL",
		m.formModel.descriptionInput,
	)
	if err != nil {
		return errMsg{err: err}
	}
	
	return accountActionMsg{}
}

func (m *AccountsModel) updateAccount(accountType entity.AccountType, balance float64) tea.Msg {
	if m.formModel.editingID == nil {
		return errMsg{err: fmt.Errorf("no account ID for editing")}
	}
	
	_, err := m.accountUseCase.UpdateAccount(
		m.ctx,
		*m.formModel.editingID,
		m.formModel.nameInput,
		accountType,
		balance,
		"BRL",
		m.formModel.descriptionInput,
	)
	if err != nil {
		return errMsg{err: err}
	}
	
	return accountActionMsg{}
}

func (m *AccountsModel) deleteAccount() tea.Msg {
	if m.selectedIndex >= len(m.accounts) {
		return errMsg{err: fmt.Errorf("no account selected")}
	}
	
	account := m.accounts[m.selectedIndex]
	err := m.accountUseCase.DeleteAccount(m.ctx, account.ID)
	if err != nil {
		return errMsg{err: err}
	}
	
	return accountActionMsg{}
}

func (m *AccountsModel) renderAccountForm() string {
	var sections []string
	
	title := "📝 New Account"
	if m.formModel.editing {
		title = "✏️ Edit Account"
	}
	sections = append(sections, style.TitleStyle.Render(title))
	
	form := m.renderForm()
	sections = append(sections, form)
	
	help := m.renderFormHelp()
	sections = append(sections, help)
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *AccountsModel) renderForm() string {
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(2, 4).
		MarginTop(1)
	
	var fields []string
	
	fields = append(fields, m.renderFormField("Name:", m.formModel.nameInput, 0))
	fields = append(fields, m.renderTypeSelector())
	fields = append(fields, m.renderFormField("Initial Balance:", m.formModel.balanceInput, 2))
	fields = append(fields, m.renderFormField("Description:", m.formModel.descriptionInput, 3))
	
	buttons := m.renderFormButtons()
	fields = append(fields, buttons)
	
	content := strings.Join(fields, "\n\n")
	return formStyle.Render(content)
}

func (m *AccountsModel) renderFormField(label, value string, fieldIndex int) string {
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

func (m *AccountsModel) renderTypeSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)
	
	var options []string
	for i, option := range m.formModel.typeOptions {
		if i == m.formModel.selectedType {
			if m.formModel.focusedField == 1 {
				options = append(options, style.SelectedMenuItemStyle.Render("► "+option))
			} else {
				options = append(options, style.InfoStyle.Render("• "+option))
			}
		} else {
			options = append(options, style.MenuItemStyle.Render("  "+option))
		}
	}
	
	selector := lipgloss.JoinHorizontal(lipgloss.Left, options...)
	
	if m.formModel.focusedField == 1 {
		selector = selector + " ◄"
	}
	
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Type:"),
		selector,
	)
}

func (m *AccountsModel) renderFormButtons() string {
	submitText := "Create Account"
	if m.formModel.editing {
		submitText = "Update Account"
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

func (m *AccountsModel) renderFormHelp() string {
	help := "[Tab] Next Field • [Shift+Tab] Previous • [←/→] Select Type • [Enter] Confirm • [Esc] Cancel"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

func (m *AccountsModel) renderConfirmDialog() string {
	if m.selectedIndex >= len(m.accounts) {
		return ""
	}
	
	account := m.accounts[m.selectedIndex]
	
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Danger).
		Padding(2, 4).
		MarginTop(5)
	
	title := style.ErrorStyle.Render("⚠️  Confirm Delete")
	message := fmt.Sprintf("Are you sure you want to delete account '%s'?", account.Name)
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

func (m *AccountsModel) IsInFormMode() bool {
	return m.viewMode == AccountViewForm || m.viewMode == AccountViewConfirm
}

type accountsLoadedMsg struct {
	accounts []*entity.Account
}

type accountActionMsg struct{}

type BackToDashboardMsg struct{}

