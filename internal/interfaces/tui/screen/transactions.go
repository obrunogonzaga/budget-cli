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

type TransactionsModel struct {
	ctx                      context.Context
	transactionUseCase       *usecase.TransactionUseCase
	accountUseCase           *usecase.AccountUseCase
	creditCardUseCase        *usecase.CreditCardUseCase
	creditCardInvoiceUseCase *usecase.CreditCardInvoiceUseCase
	billUseCase              *usecase.BillUseCase
	personUseCase            *usecase.PersonUseCase

	// Data
	transactions         []*entity.Transaction
	filteredTransactions []*entity.Transaction
	accounts             []*entity.Account
	creditCards          []*entity.CreditCard
	people               []*entity.Person

	// View state
	selectedIndex int
	viewMode      TransactionViewMode

	// Pagination
	currentPage  int
	itemsPerPage int

	// Loading and errors
	loading bool
	err     error

	// Form state
	formModel *TransactionFormModel

	// Filter state
	filterModel *TransactionFilterModel

	// Shared expense state
	sharedModel *SharedExpenseModel

	// Invoice state
	invoiceModel *InvoiceViewModel

	// Confirmation state
	showConfirmDelete bool
	confirmMessage    string

	// Window dimensions
	width  int
	height int
}

type TransactionViewMode int

const (
	TransactionViewList TransactionViewMode = iota
	TransactionViewForm
	TransactionViewDetails
	TransactionViewShared
	TransactionViewFilter
	TransactionViewConfirm
	TransactionViewInvoices
	TransactionViewInvoiceTransactions
)

type TransactionFormModel struct {
	// Form fields
	description string
	amount      string
	date        string

	// Selection states
	selectedType     int // 0: expense, 1: income
	selectedCategory int
	selectedSource   int // 0: account, 1: credit card
	selectedAccount  int
	selectedCard     int

	// Input fields
	descriptionInput string
	amountInput      string
	dateInput        string

	// Navigation
	focusedField int

	// Edit state
	editing   bool
	editingID *uuid.UUID
}

type TransactionFilterModel struct {
	// Date range
	dateRangeType int // 0: all, 1: today, 2: this week, 3: this month, 4: custom
	startDate     string
	endDate       string

	// Category filter
	selectedCategories map[entity.TransactionCategory]bool

	// Account/Card filter
	filterBySource    int // 0: all, 1: accounts only, 2: cards only
	selectedAccountID *uuid.UUID
	selectedCardID    *uuid.UUID

	// Type filter
	typeFilter int // 0: all, 1: income only, 2: expense only

	// Navigation
	focusedSection int
	focusedField   int
}

type SharedExpenseModel struct {
	transactionID uuid.UUID
	transaction   *entity.Transaction

	// Split type
	splitType int // 0: equal, 1: custom

	// People selection
	selectedPeople map[uuid.UUID]bool
	customAmounts  map[uuid.UUID]string

	// Add person mode
	addingPerson   bool
	newPersonName  string
	newPersonEmail string

	// Navigation
	focusedField int
}

type InvoiceViewModel struct {
	// Invoice data
	invoices []*entity.CreditCardInvoice
	invoiceTransactions []*entity.Transaction
	selectedInvoice *entity.CreditCardInvoice

	// Navigation
	selectedInvoiceIndex int
	selectedTransactionIndex int

	// Pagination for transactions
	currentTransactionPage int
}

func NewTransactionsModel(ctx context.Context, txnUC *usecase.TransactionUseCase, accountUC *usecase.AccountUseCase, cardUC *usecase.CreditCardUseCase, invoiceUC *usecase.CreditCardInvoiceUseCase, billUC *usecase.BillUseCase, personUC *usecase.PersonUseCase) tea.Model {
	return &TransactionsModel{
		ctx:                      ctx,
		transactionUseCase:       txnUC,
		accountUseCase:           accountUC,
		creditCardUseCase:        cardUC,
		creditCardInvoiceUseCase: invoiceUC,
		billUseCase:              billUC,
		personUseCase:            personUC,
		viewMode:                 TransactionViewList,
		loading:                  true,
		itemsPerPage:             10,
		currentPage:              0,
		formModel: &TransactionFormModel{
			date:      time.Now().Format("2006-01-02"),
			dateInput: time.Now().Format("2006-01-02"),
		},
		filterModel: &TransactionFilterModel{
			selectedCategories: make(map[entity.TransactionCategory]bool),
			dateRangeType:      0, // All
		},
		sharedModel: &SharedExpenseModel{
			selectedPeople: make(map[uuid.UUID]bool),
			customAmounts:  make(map[uuid.UUID]string),
		},
		invoiceModel: &InvoiceViewModel{
			invoices: []*entity.CreditCardInvoice{},
			invoiceTransactions: []*entity.Transaction{},
			selectedInvoiceIndex: 0,
			selectedTransactionIndex: 0,
			currentTransactionPage: 0,
		},
	}
}

func (m *TransactionsModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadTransactions,
		m.loadAccounts,
		m.loadCreditCards,
		m.loadPeople,
	)
}

func (m *TransactionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case transactionsLoadedMsg:
		m.loading = false
		m.transactions = msg.transactions
		m.applyFilters()
		return m, nil

	case accountsLoadedMsg:
		m.accounts = msg.accounts
		return m, nil

	case creditCardsLoadedMsg:
		m.creditCards = msg.creditCards
		return m, nil

	case peopleLoadedMsg:
		m.people = msg.people
		return m, nil

	case transactionActionMsg:
		m.loading = false
		m.viewMode = TransactionViewList
		m.resetForm()
		m.resetSharedModel()
		return m, m.loadTransactions

	case invoicesLoadedMsg:
		m.loading = false
		m.invoiceModel.invoices = msg.invoices
		return m, nil

	case invoiceTransactionsLoadedMsg:
		m.loading = false
		m.invoiceModel.invoiceTransactions = msg.transactions
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch m.viewMode {
		case TransactionViewList:
			return m.handleListKeys(msg)
		case TransactionViewForm:
			return m.handleFormKeys(msg)
		case TransactionViewDetails:
			return m.handleDetailsKeys(msg)
		case TransactionViewShared:
			return m.handleSharedKeys(msg)
		case TransactionViewFilter:
			return m.handleFilterKeys(msg)
		case TransactionViewConfirm:
			return m.handleConfirmKeys(msg)
		case TransactionViewInvoices:
			return m.handleInvoicesKeys(msg)
		case TransactionViewInvoiceTransactions:
			return m.handleInvoiceTransactionsKeys(msg)
		}
	}

	return m, nil
}

func (m *TransactionsModel) View() string {
	if m.loading {
		return style.InfoStyle.Render("Loading transactions...")
	}

	if m.err != nil {
		return style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.viewMode {
	case TransactionViewList:
		return m.renderTransactionsList()
	case TransactionViewForm:
		return m.renderTransactionForm()
	case TransactionViewDetails:
		return m.renderTransactionDetails()
	case TransactionViewShared:
		return m.renderSharedExpense()
	case TransactionViewFilter:
		return m.renderFilterView()
	case TransactionViewConfirm:
		return m.renderConfirmDialog()
	case TransactionViewInvoices:
		return m.renderInvoicesList()
	case TransactionViewInvoiceTransactions:
		return m.renderInvoiceTransactions()
	}

	return ""
}

// Helper functions for loading data
func (m *TransactionsModel) loadTransactions() tea.Msg {
	transactions, err := m.transactionUseCase.GetTransactionsByDateRange(
		m.ctx,
		time.Now().AddDate(-1, 0, 0), // Last year
		time.Now().AddDate(0, 0, 1),  // Tomorrow
	)
	if err != nil {
		return errMsg{err: err}
	}

	return transactionsLoadedMsg{transactions: transactions}
}

func (m *TransactionsModel) loadAccounts() tea.Msg {
	accounts, err := m.accountUseCase.ListAccounts(m.ctx)
	if err != nil {
		// Don't fail the whole screen if accounts can't be loaded
		return accountsLoadedMsg{accounts: []*entity.Account{}}
	}
	return accountsLoadedMsg{accounts: accounts}
}

func (m *TransactionsModel) loadCreditCards() tea.Msg {
	cards, err := m.creditCardUseCase.ListCreditCards(m.ctx)
	if err != nil {
		// Don't fail the whole screen if cards can't be loaded
		return creditCardsLoadedMsg{creditCards: []*entity.CreditCard{}}
	}
	return creditCardsLoadedMsg{creditCards: cards}
}

func (m *TransactionsModel) loadPeople() tea.Msg {
	people, err := m.personUseCase.ListPeople(m.ctx)
	if err != nil {
		// Don't fail the whole screen if people can't be loaded
		return peopleLoadedMsg{people: []*entity.Person{}}
	}
	return peopleLoadedMsg{people: people}
}

// Filter transactions based on current filter settings
func (m *TransactionsModel) applyFilters() {
	filtered := make([]*entity.Transaction, 0)

	for _, txn := range m.transactions {
		// Date filter
		if !m.matchesDateFilter(txn) {
			continue
		}

		// Category filter
		if len(m.filterModel.selectedCategories) > 0 && !m.filterModel.selectedCategories[txn.Category] {
			continue
		}

		// Source filter
		if !m.matchesSourceFilter(txn) {
			continue
		}

		// Type filter
		if !m.matchesTypeFilter(txn) {
			continue
		}

		filtered = append(filtered, txn)
	}

	m.filteredTransactions = filtered
	m.currentPage = 0
	m.selectedIndex = 0
}

func (m *TransactionsModel) matchesDateFilter(txn *entity.Transaction) bool {
	switch m.filterModel.dateRangeType {
	case 0: // All
		return true
	case 1: // Today
		today := time.Now().Truncate(24 * time.Hour)
		txnDate := txn.Date.Truncate(24 * time.Hour)
		return txnDate.Equal(today)
	case 2: // This week
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday()))
		weekEnd := weekStart.AddDate(0, 0, 7)
		return txn.Date.After(weekStart) && txn.Date.Before(weekEnd)
	case 3: // This month
		now := time.Now()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)
		return txn.Date.After(monthStart) && txn.Date.Before(monthEnd)
	case 4: // Custom
		// Parse custom date range
		startDate, _ := time.Parse("2006-01-02", m.filterModel.startDate)
		endDate, _ := time.Parse("2006-01-02", m.filterModel.endDate)
		return txn.Date.After(startDate) && txn.Date.Before(endDate.AddDate(0, 0, 1))
	}
	return true
}

func (m *TransactionsModel) matchesSourceFilter(txn *entity.Transaction) bool {
	switch m.filterModel.filterBySource {
	case 0: // All
		return true
	case 1: // Accounts only
		if txn.AccountID == nil {
			return false
		}
		if m.filterModel.selectedAccountID != nil {
			return *txn.AccountID == *m.filterModel.selectedAccountID
		}
		return true
	case 2: // Cards only
		if txn.CreditCardID == nil {
			return false
		}
		if m.filterModel.selectedCardID != nil {
			return *txn.CreditCardID == *m.filterModel.selectedCardID
		}
		return true
	}
	return true
}

func (m *TransactionsModel) matchesTypeFilter(txn *entity.Transaction) bool {
	switch m.filterModel.typeFilter {
	case 0: // All
		return true
	case 1: // Income only
		return txn.Type == entity.TransactionTypeCredit
	case 2: // Expense only
		return txn.Type == entity.TransactionTypeDebit
	}
	return true
}

// Helper to reset form
func (m *TransactionsModel) resetForm() {
	m.formModel = &TransactionFormModel{
		date:             time.Now().Format("2006-01-02"),
		dateInput:        time.Now().Format("2006-01-02"),
		focusedField:     0,
		selectedType:     0,
		selectedCategory: 0,
		selectedSource:   0,
		selectedAccount:  0,
		selectedCard:     0,
	}
}

// Helper to reset shared expense model
func (m *TransactionsModel) resetSharedModel() {
	m.sharedModel = &SharedExpenseModel{
		selectedPeople: make(map[uuid.UUID]bool),
		customAmounts:  make(map[uuid.UUID]string),
	}
}

// FormModeChecker interface implementation
func (m *TransactionsModel) IsInFormMode() bool {
	return m.viewMode != TransactionViewList
}

// Message types
type transactionsLoadedMsg struct {
	transactions []*entity.Transaction
}

type creditCardsLoadedMsg struct {
	creditCards []*entity.CreditCard
}

type peopleLoadedMsg struct {
	people []*entity.Person
}

type transactionActionMsg struct{}



// Key handler for list view
func (m *TransactionsModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalPages := (len(m.filteredTransactions) + m.itemsPerPage - 1) / m.itemsPerPage
	pageStart := m.currentPage * m.itemsPerPage
	pageEnd := pageStart + m.itemsPerPage
	if pageEnd > len(m.filteredTransactions) {
		pageEnd = len(m.filteredTransactions)
	}
	itemsOnPage := pageEnd - pageStart

	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		} else if m.currentPage > 0 {
			// Move to previous page
			m.currentPage--
			m.selectedIndex = m.itemsPerPage - 1
		}
	case "down", "j":
		if m.selectedIndex < itemsOnPage-1 {
			m.selectedIndex++
		} else if m.currentPage < totalPages-1 {
			// Move to next page
			m.currentPage++
			m.selectedIndex = 0
		}
	case "left", "h":
		if m.currentPage > 0 {
			m.currentPage--
			m.selectedIndex = 0
		}
	case "right", "l":
		if m.currentPage < totalPages-1 {
			m.currentPage++
			m.selectedIndex = 0
		}
	case "enter":
		if len(m.filteredTransactions) > 0 {
			m.viewMode = TransactionViewDetails
		}
	case "n":
		m.viewMode = TransactionViewForm
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
	case "e":
		if len(m.filteredTransactions) > 0 {
			return m.editTransaction()
		}
	case "d":
		if len(m.filteredTransactions) > 0 {
			idx := m.currentPage*m.itemsPerPage + m.selectedIndex
			if idx < len(m.filteredTransactions) {
				m.viewMode = TransactionViewConfirm
				m.showConfirmDelete = true
			}
		}
	case "s":
		if len(m.filteredTransactions) > 0 {
			idx := m.currentPage*m.itemsPerPage + m.selectedIndex
			if idx < len(m.filteredTransactions) {
				txn := m.filteredTransactions[idx]
				m.sharedModel.transactionID = txn.ID
				m.sharedModel.transaction = txn
				m.viewMode = TransactionViewShared
			}
		}
	case "f":
		m.viewMode = TransactionViewFilter
	case "i":
		m.viewMode = TransactionViewInvoices
		m.loading = true
		return m, m.loadAllInvoices
	case "r":
		m.loading = true
		return m, tea.Batch(
			m.loadTransactions,
			m.loadAccounts,
			m.loadCreditCards,
			m.loadPeople,
		)
	case "b":
		return m, func() tea.Msg { return BackToDashboardMsg{} }
	}

	return m, nil
}

func (m *TransactionsModel) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalFields := 8 // description, type, category, amount, date, source, account/card, buttons

	switch msg.String() {
	case "esc":
		m.viewMode = TransactionViewList
		m.resetForm()
	case "tab", "down":
		m.formModel.focusedField = (m.formModel.focusedField + 1) % totalFields
	case "shift+tab", "up":
		m.formModel.focusedField = (m.formModel.focusedField - 1 + totalFields) % totalFields
	case "enter":
		if m.formModel.focusedField == 6 {
			// Submit button
			return m.submitForm()
		} else if m.formModel.focusedField == 7 {
			// Cancel button
			m.viewMode = TransactionViewList
			m.resetForm()
		}
	default:
		return m.handleFormInput(msg)
	}

	return m, nil
}

func (m *TransactionsModel) handleDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "b", "enter":
		m.viewMode = TransactionViewList
	case "e":
		return m.editTransaction()
	case "d":
		m.viewMode = TransactionViewConfirm
		m.showConfirmDelete = true
	case "s":
		idx := m.currentPage*m.itemsPerPage + m.selectedIndex
		if idx < len(m.filteredTransactions) {
			txn := m.filteredTransactions[idx]
			m.sharedModel.transactionID = txn.ID
			m.sharedModel.transaction = txn
			m.viewMode = TransactionViewShared
		}
	}

	return m, nil
}

func (m *TransactionsModel) handleSharedKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement shared expense key handling
	return m, nil
}

func (m *TransactionsModel) handleFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement filter key handling
	return m, nil
}

func (m *TransactionsModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.viewMode = TransactionViewList
		m.showConfirmDelete = false
		m.loading = true
		return m, m.deleteTransaction
	case "n", "esc":
		// If we came from details view, go back there
		if m.viewMode == TransactionViewConfirm && !m.showConfirmDelete {
			m.viewMode = TransactionViewDetails
		} else {
			m.viewMode = TransactionViewList
		}
		m.showConfirmDelete = false
	}

	return m, nil
}

// Helper method to edit a transaction
func (m *TransactionsModel) editTransaction() (tea.Model, tea.Cmd) {
	idx := m.currentPage*m.itemsPerPage + m.selectedIndex
	if idx >= len(m.filteredTransactions) {
		return m, nil
	}

	txn := m.filteredTransactions[idx]
	m.viewMode = TransactionViewForm
	m.formModel.editing = true
	m.formModel.editingID = &txn.ID

	// Pre-fill form with transaction data
	m.formModel.descriptionInput = txn.Description
	m.formModel.amountInput = fmt.Sprintf("%.2f", txn.Amount.Amount())
	m.formModel.dateInput = txn.Date.Format("2006-01-02")

	// Set type
	if txn.Type == entity.TransactionTypeCredit {
		m.formModel.selectedType = 1
	} else {
		m.formModel.selectedType = 0
	}

	// Set category
	categories := m.getCategories()
	for i, cat := range categories {
		if cat == txn.Category {
			m.formModel.selectedCategory = i
			break
		}
	}

	// Set source
	if txn.AccountID != nil {
		m.formModel.selectedSource = 0
		for i, acc := range m.accounts {
			if acc.ID == *txn.AccountID {
				m.formModel.selectedAccount = i
				break
			}
		}
	} else if txn.CreditCardID != nil {
		m.formModel.selectedSource = 1
		for i, card := range m.creditCards {
			if card.ID == *txn.CreditCardID {
				m.formModel.selectedCard = i
				break
			}
		}
	}

	return m, nil
}

// Get list of transaction categories
func (m *TransactionsModel) getCategories() []entity.TransactionCategory {
	return []entity.TransactionCategory{
		entity.TransactionCategoryFood,
		entity.TransactionCategoryTransportation,
		entity.TransactionCategoryUtilities,
		entity.TransactionCategoryEntertainment,
		entity.TransactionCategoryShopping,
		entity.TransactionCategoryHealthcare,
		entity.TransactionCategoryEducation,
		entity.TransactionCategoryIncome,
		entity.TransactionCategoryTransfer,
		entity.TransactionCategoryOther,
	}
}

// Get category display name
func (m *TransactionsModel) getCategoryDisplay(cat entity.TransactionCategory) string {
	switch cat {
	case entity.TransactionCategoryFood:
		return "🍔 Food"
	case entity.TransactionCategoryTransportation:
		return "🚗 Transportation"
	case entity.TransactionCategoryUtilities:
		return "💡 Utilities"
	case entity.TransactionCategoryEntertainment:
		return "🎮 Entertainment"
	case entity.TransactionCategoryShopping:
		return "🛍️ Shopping"
	case entity.TransactionCategoryHealthcare:
		return "🏥 Healthcare"
	case entity.TransactionCategoryEducation:
		return "📚 Education"
	case entity.TransactionCategoryIncome:
		return "💰 Income"
	case entity.TransactionCategoryTransfer:
		return "🔄 Transfer"
	case entity.TransactionCategoryOther:
		return "📋 Other"
	default:
		return string(cat)
	}
}

// View rendering for transaction list
func (m *TransactionsModel) renderTransactionsList() string {
	var sections []string

	title := style.TitleStyle.Render("💸 Transactions Management")
	sections = append(sections, title)

	// Summary bar
	summary := m.renderSummaryBar()
	sections = append(sections, summary)

	if len(m.filteredTransactions) == 0 {
		empty := style.InfoStyle.Render("No transactions found. Press 'n' to create your first transaction.")
		sections = append(sections, empty)
	} else {
		// Transaction table
		table := m.renderTransactionsTable()
		sections = append(sections, table)

		// Pagination info
		pagination := m.renderPagination()
		sections = append(sections, pagination)
	}

	help := m.renderListHelp()
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render summary bar with totals
func (m *TransactionsModel) renderSummaryBar() string {
	var totalIncome, totalExpense float64

	for _, txn := range m.filteredTransactions {
		if txn.Type == entity.TransactionTypeCredit {
			totalIncome += txn.Amount.Amount()
		} else {
			totalExpense += txn.Amount.Amount()
		}
	}

	balance := totalIncome - totalExpense

	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(0, 1).
		MarginTop(1)

	incomeStr := style.SuccessStyle.Render(fmt.Sprintf("Income: R$ %.2f", totalIncome))
	expenseStr := style.ErrorStyle.Render(fmt.Sprintf("Expense: R$ %.2f", totalExpense))

	var balanceStr string
	if balance >= 0 {
		balanceStr = style.SuccessStyle.Render(fmt.Sprintf("Balance: R$ %.2f", balance))
	} else {
		balanceStr = style.ErrorStyle.Render(fmt.Sprintf("Balance: R$ %.2f", balance))
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		incomeStr,
		"  |  ",
		expenseStr,
		"  |  ",
		balanceStr,
	)

	return summaryStyle.Render(content)
}

// Render transactions table
func (m *TransactionsModel) renderTransactionsTable() string {
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)

	headers := []string{"Date", "Description", "Category", "Amount", "Source", "Shared"}
	headerRow := style.TableHeaderStyle.Render(
		fmt.Sprintf("%-12s %-25s %-15s %-12s %-15s %-6s",
			headers[0], headers[1], headers[2], headers[3], headers[4], headers[5]),
	)

	var rows []string
	rows = append(rows, headerRow)

	// Calculate page boundaries
	start := m.currentPage * m.itemsPerPage
	end := start + m.itemsPerPage
	if end > len(m.filteredTransactions) {
		end = len(m.filteredTransactions)
	}

	for i := start; i < end; i++ {
		txn := m.filteredTransactions[i]

		date := txn.Date.Format("2006-01-02")
		description := truncateString(txn.Description, 25)
		category := truncateString(m.getCategoryDisplay(txn.Category), 15)

		// Format amount with color
		var amountStr string
		amount := fmt.Sprintf("R$ %.2f", txn.Amount.Amount())
		if txn.Type == entity.TransactionTypeCredit {
			amountStr = style.SuccessStyle.Render("+" + amount)
		} else {
			amountStr = style.ErrorStyle.Render("-" + amount)
		}

		// Get source
		source := m.getTransactionSource(txn)
		source = truncateString(source, 15)

		// Check if shared
		shared := ""
		if len(txn.SharedWith) > 0 {
			shared = "👥"
		}

		row := fmt.Sprintf("%-12s %-25s %-15s %-12s %-15s %-6s",
			date, description, category, amountStr, source, shared)

		if i-start == m.selectedIndex {
			row = style.SelectedMenuItemStyle.Render("► " + row)
		} else {
			row = style.MenuItemStyle.Render("  " + row)
		}

		rows = append(rows, row)
	}

	table := strings.Join(rows, "\n")
	return tableStyle.Render(table)
}

// Get transaction source display
func (m *TransactionsModel) getTransactionSource(txn *entity.Transaction) string {
	if txn.AccountID != nil {
		for _, acc := range m.accounts {
			if acc.ID == *txn.AccountID {
				return acc.Name
			}
		}
		return "Unknown Account"
	} else if txn.CreditCardID != nil {
		for _, card := range m.creditCards {
			if card.ID == *txn.CreditCardID {
				return card.Name
			}
		}
		return "Unknown Card"
	}
	return "Cash"
}

// Render pagination information
func (m *TransactionsModel) renderPagination() string {
	totalPages := (len(m.filteredTransactions) + m.itemsPerPage - 1) / m.itemsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	paginationStyle := lipgloss.NewStyle().
		Foreground(style.TextMuted).
		MarginTop(1)

	info := fmt.Sprintf("Page %d of %d | Total: %d transactions | Use ← → to navigate pages",
		m.currentPage+1, totalPages, len(m.filteredTransactions))

	return paginationStyle.Render(info)
}

// Render help text for list view
func (m *TransactionsModel) renderListHelp() string {
	help := "[↑/↓] Navigate • [Enter] Details • [n] New • [e] Edit • [d] Delete • [s] Share • [f] Filter • [i] Invoices • [r] Refresh • [b] Back"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

// Handle form input based on focused field
func (m *TransactionsModel) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.formModel.focusedField {
	case 0: // Description
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
	case 1: // Type (income/expense)
		switch msg.String() {
		case "left":
			m.formModel.selectedType = 0
		case "right":
			m.formModel.selectedType = 1
		}
	case 2: // Category
		categories := m.getCategories()
		switch msg.String() {
		case "left":
			if m.formModel.selectedCategory > 0 {
				m.formModel.selectedCategory--
			}
		case "right":
			if m.formModel.selectedCategory < len(categories)-1 {
				m.formModel.selectedCategory++
			}
		}
	case 3: // Amount
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
	case 4: // Date
		switch msg.String() {
		case "backspace":
			if len(m.formModel.dateInput) > 0 {
				m.formModel.dateInput = m.formModel.dateInput[:len(m.formModel.dateInput)-1]
			}
		default:
			if len(msg.String()) == 1 && (msg.String() >= "0" && msg.String() <= "9" || msg.String() == "-") {
				m.formModel.dateInput += msg.String()
			}
		}
	case 5: // Source type (account/card)
		switch msg.String() {
		case "left":
			m.formModel.selectedSource = 0
		case "right":
			m.formModel.selectedSource = 1
		}
		// Reset selection when changing source type
		m.formModel.selectedAccount = 0
		m.formModel.selectedCard = 0
	case 6: // Account or Card selection
		if m.formModel.selectedSource == 0 {
			// Account selection
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
		} else {
			// Card selection
			switch msg.String() {
			case "left":
				if m.formModel.selectedCard > 0 {
					m.formModel.selectedCard--
				}
			case "right":
				if m.formModel.selectedCard < len(m.creditCards)-1 {
					m.formModel.selectedCard++
				}
			}
		}
	}

	return m, nil
}

// Submit transaction form
func (m *TransactionsModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate form
	if strings.TrimSpace(m.formModel.descriptionInput) == "" {
		m.err = fmt.Errorf("description is required")
		return m, nil
	}

	amount, err := strconv.ParseFloat(m.formModel.amountInput, 64)
	if err != nil || amount <= 0 {
		m.err = fmt.Errorf("invalid amount")
		return m, nil
	}

	date, err := time.Parse("2006-01-02", m.formModel.dateInput)
	if err != nil {
		m.err = fmt.Errorf("invalid date format (use YYYY-MM-DD)")
		return m, nil
	}

	// Get transaction type
	var txnType entity.TransactionType
	if m.formModel.selectedType == 0 {
		txnType = entity.TransactionTypeDebit
	} else {
		txnType = entity.TransactionTypeCredit
	}

	// Get category
	categories := m.getCategories()
	category := categories[m.formModel.selectedCategory]

	// Get account/card
	var accountID *uuid.UUID
	var creditCardID *uuid.UUID

	if m.formModel.selectedSource == 0 && len(m.accounts) > 0 {
		accountID = &m.accounts[m.formModel.selectedAccount].ID
	} else if m.formModel.selectedSource == 1 && len(m.creditCards) > 0 {
		creditCardID = &m.creditCards[m.formModel.selectedCard].ID
	}

	m.loading = true

	if m.formModel.editing && m.formModel.editingID != nil {
		// TODO: Implement update transaction
		return m, func() tea.Msg {
			return errMsg{err: fmt.Errorf("update transaction not implemented yet")}
		}
	}

	// Create transaction
	return m, func() tea.Msg {
		_, err := m.transactionUseCase.CreateTransaction(
			m.ctx,
			accountID,
			creditCardID,
			txnType,
			category,
			amount,
			"BRL",
			m.formModel.descriptionInput,
			date,
		)
		if err != nil {
			return errMsg{err: err}
		}

		return transactionActionMsg{}
	}
}

func (m *TransactionsModel) renderTransactionForm() string {
	var sections []string

	title := "📝 New Transaction"
	if m.formModel.editing {
		title = "✏️ Edit Transaction"
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

// Render the transaction form
func (m *TransactionsModel) renderForm() string {
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(2, 4).
		MarginTop(1)

	var fields []string

	// Description field
	fields = append(fields, m.renderFormField("Description:", m.formModel.descriptionInput, 0))

	// Type selector (Income/Expense)
	fields = append(fields, m.renderTypeSelector())

	// Category selector
	fields = append(fields, m.renderCategorySelector())

	// Amount field
	fields = append(fields, m.renderFormField("Amount (R$):", m.formModel.amountInput, 3))

	// Date field
	fields = append(fields, m.renderFormField("Date:", m.formModel.dateInput, 4))

	// Source selector (Account/Card)
	fields = append(fields, m.renderSourceSelector())

	// Account/Card selector
	if m.formModel.selectedSource == 0 {
		fields = append(fields, m.renderAccountSelector())
	} else {
		fields = append(fields, m.renderCardSelector())
	}

	// Buttons
	buttons := m.renderFormButtons()
	fields = append(fields, buttons)

	content := strings.Join(fields, "\n\n")
	return formStyle.Render(content)
}

// Render a form field
func (m *TransactionsModel) renderFormField(label, value string, fieldIndex int) string {
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

// Render type selector (Income/Expense)
func (m *TransactionsModel) renderTypeSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)

	types := []string{"💸 Expense", "💰 Income"}
	var options []string

	for i, t := range types {
		if i == m.formModel.selectedType {
			if m.formModel.focusedField == 1 {
				options = append(options, style.SelectedMenuItemStyle.Render("► "+t))
			} else {
				options = append(options, style.InfoStyle.Render("• "+t))
			}
		} else {
			options = append(options, style.MenuItemStyle.Render("  "+t))
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

// Render category selector
func (m *TransactionsModel) renderCategorySelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)

	categories := m.getCategories()
	selectedCat := categories[m.formModel.selectedCategory]
	display := m.getCategoryDisplay(selectedCat)

	var selector string
	if m.formModel.focusedField == 2 {
		selector = style.FocusedInputStyle.Width(30).Render("< " + display + " >")
		selector = selector + " ◄"
	} else {
		selector = style.InputStyle.Width(30).Render(display)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Category:"),
		selector,
	)
}

// Render source selector (Account/Card)
func (m *TransactionsModel) renderSourceSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)

	sources := []string{"🏦 Account", "💳 Credit Card"}
	var options []string

	for i, s := range sources {
		if i == m.formModel.selectedSource {
			if m.formModel.focusedField == 5 {
				options = append(options, style.SelectedMenuItemStyle.Render("► "+s))
			} else {
				options = append(options, style.InfoStyle.Render("• "+s))
			}
		} else {
			options = append(options, style.MenuItemStyle.Render("  "+s))
		}
	}

	selector := lipgloss.JoinHorizontal(lipgloss.Left, options...)

	if m.formModel.focusedField == 5 {
		selector = selector + " ◄"
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Source:"),
		selector,
	)
}

// Render account selector
func (m *TransactionsModel) renderAccountSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)

	if len(m.accounts) == 0 {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Account:"),
			style.WarningStyle.Render("No accounts available"),
		)
	}

	account := m.accounts[m.formModel.selectedAccount]
	display := fmt.Sprintf("%s (R$ %.2f)", account.Name, account.Balance.Amount())

	var selector string
	if m.formModel.focusedField == 6 {
		selector = style.FocusedInputStyle.Width(30).Render("< " + display + " >")
		selector = selector + " ◄"
	} else {
		selector = style.InputStyle.Width(30).Render(display)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Account:"),
		selector,
	)
}

// Render credit card selector
func (m *TransactionsModel) renderCardSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		Width(20)

	if len(m.creditCards) == 0 {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Credit Card:"),
			style.WarningStyle.Render("No credit cards available"),
		)
	}

	card := m.creditCards[m.formModel.selectedCard]
	available := card.CreditLimit.Amount() - card.CurrentBalance.Amount()
	display := fmt.Sprintf("%s (Available: R$ %.2f)", card.Name, available)

	var selector string
	if m.formModel.focusedField == 6 {
		selector = style.FocusedInputStyle.Width(30).Render("< " + display + " >")
		selector = selector + " ◄"
	} else {
		selector = style.InputStyle.Width(30).Render(display)
	}

	// Get invoice info for the selected date
	invoiceInfo := m.getInvoiceInfo(card.ID)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render("Credit Card:"),
			selector,
		),
		invoiceInfo,
	)
}

// Get invoice info for the selected date and card
func (m *TransactionsModel) getInvoiceInfo(cardID uuid.UUID) string {
	// Parse the selected date
	date, err := time.Parse("2006-01-02", m.formModel.date)
	if err != nil {
		return style.ErrorStyle.Render("  ⚠ Invalid date format")
	}

	// Get invoices for the credit card
	invoices, err := m.creditCardInvoiceUseCase.ListInvoicesByCard(m.ctx, cardID)
	if err != nil {
		return style.ErrorStyle.Render("  ⚠ Failed to load invoices")
	}

	// Find the invoice that contains this date
	var targetInvoice *entity.CreditCardInvoice
	for _, invoice := range invoices {
		if invoice.ContainsDate(date) {
			targetInvoice = invoice
			break
		}
	}

	// If no invoice found, create preview for new invoice
	if targetInvoice == nil {
		year, month := date.Year(), date.Month()
		referenceMonth := fmt.Sprintf("%04d-%02d", year, month)

		return style.InfoStyle.Render(
			fmt.Sprintf("  📋 Will create new invoice for %s", referenceMonth),
		)
	}

	// Display current invoice info
	statusStyle := style.SuccessStyle
	if targetInvoice.Status == entity.InvoiceStatusClosed {
		statusStyle = style.WarningStyle
	} else if targetInvoice.Status == entity.InvoiceStatusOverdue {
		statusStyle = style.ErrorStyle
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.InfoStyle.Render(fmt.Sprintf("  📋 Invoice: %s", targetInvoice.ReferenceMonth)),
		statusStyle.Render(fmt.Sprintf("  Status: %s", targetInvoice.Status)),
		style.InfoStyle.Render(fmt.Sprintf("  Due: %s", targetInvoice.DueDate.Format("2006-01-02"))),
	)
}

// Render form buttons
func (m *TransactionsModel) renderFormButtons() string {
	submitText := "Create Transaction"
	if m.formModel.editing {
		submitText = "Update Transaction"
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

// Render form help
func (m *TransactionsModel) renderFormHelp() string {
	help := "[Tab] Next Field • [Shift+Tab] Previous • [←/→] Select Option • [Enter] Confirm • [Esc] Cancel"
	return style.HelpStyle.
		MarginTop(1).
		Render(help)
}

// Delete transaction
func (m *TransactionsModel) deleteTransaction() tea.Msg {
	idx := m.currentPage*m.itemsPerPage + m.selectedIndex
	if idx >= len(m.filteredTransactions) {
		return errMsg{err: fmt.Errorf("no transaction selected")}
	}

	_ = m.filteredTransactions[idx]

	// TODO: Implement delete in use case
	// For now, we just return an error
	return errMsg{err: fmt.Errorf("delete transaction not implemented yet")}
}

func (m *TransactionsModel) renderTransactionDetails() string {
	idx := m.currentPage*m.itemsPerPage + m.selectedIndex
	if idx >= len(m.filteredTransactions) {
		return style.ErrorStyle.Render("No transaction selected")
	}

	txn := m.filteredTransactions[idx]

	var sections []string

	title := style.TitleStyle.Render("📋 Transaction Details")
	sections = append(sections, title)

	detailsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)

	var details []string

	// Basic info
	details = append(details, style.HeaderStyle.Render("Basic Information"))
	details = append(details, fmt.Sprintf("ID: %s", txn.ID.String()))
	details = append(details, fmt.Sprintf("Date: %s", txn.Date.Format("Monday, January 2, 2006")))
	details = append(details, fmt.Sprintf("Description: %s", txn.Description))

	// Type and amount
	details = append(details, "")
	details = append(details, style.HeaderStyle.Render("Financial Details"))
	typeStr := "Expense"
	if txn.Type == entity.TransactionTypeCredit {
		typeStr = "Income"
	}
	details = append(details, fmt.Sprintf("Type: %s", typeStr))

	amountStr := fmt.Sprintf("R$ %.2f", txn.Amount.Amount())
	if txn.Type == entity.TransactionTypeCredit {
		amountStr = style.SuccessStyle.Render("+" + amountStr)
	} else {
		amountStr = style.ErrorStyle.Render("-" + amountStr)
	}
	details = append(details, fmt.Sprintf("Amount: %s", amountStr))
	details = append(details, fmt.Sprintf("Category: %s", m.getCategoryDisplay(txn.Category)))

	// Source
	details = append(details, "")
	details = append(details, style.HeaderStyle.Render("Payment Source"))
	source := m.getTransactionSource(txn)
	details = append(details, fmt.Sprintf("Source: %s", source))

	// Shared expenses
	if len(txn.SharedWith) > 0 {
		details = append(details, "")
		details = append(details, style.HeaderStyle.Render("Shared Expenses"))
		personalAmount := txn.GetPersonalAmount()
		details = append(details, fmt.Sprintf("Your portion: R$ %.2f", personalAmount.Amount()))
		details = append(details, fmt.Sprintf("Shared with %d people", len(txn.SharedWith)))

		for _, share := range txn.SharedWith {
			personName := m.getPersonName(share.PersonID)
			details = append(details, fmt.Sprintf("  • %s: R$ %.2f (%.1f%%)",
				personName, share.Amount.Amount(), share.Percentage))
		}
	}

	// Timestamps
	details = append(details, "")
	details = append(details, style.HeaderStyle.Render("Timestamps"))
	details = append(details, fmt.Sprintf("Created: %s", txn.CreatedAt.Format("2006-01-02 15:04:05")))
	details = append(details, fmt.Sprintf("Updated: %s", txn.UpdatedAt.Format("2006-01-02 15:04:05")))

	content := strings.Join(details, "\n")
	sections = append(sections, detailsStyle.Render(content))

	help := "[Esc/Enter] Back • [e] Edit • [d] Delete • [s] Share"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *TransactionsModel) renderSharedExpense() string {
	// TODO: Implement shared expense view rendering
	return "Shared expense view - implementation pending"
}

// Get person name by ID
func (m *TransactionsModel) getPersonName(personID uuid.UUID) string {
	for _, person := range m.people {
		if person.ID == personID {
			return person.Name
		}
	}
	return "Unknown Person"
}

func (m *TransactionsModel) renderFilterView() string {
	// TODO: Implement filter view rendering
	return "Filter view - implementation pending"
}

func (m *TransactionsModel) renderConfirmDialog() string {
	idx := m.currentPage*m.itemsPerPage + m.selectedIndex
	if idx >= len(m.filteredTransactions) {
		return style.ErrorStyle.Render("No transaction selected")
	}

	txn := m.filteredTransactions[idx]

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Danger).
		Padding(2, 4).
		MarginTop(5)

	title := style.ErrorStyle.Render("⚠️  Confirm Delete")

	amountStr := fmt.Sprintf("R$ %.2f", txn.Amount.Amount())
	if txn.Type == entity.TransactionTypeCredit {
		amountStr = "+" + amountStr
	} else {
		amountStr = "-" + amountStr
	}

	message := fmt.Sprintf("Are you sure you want to delete this transaction?\n\n%s\n%s\n%s",
		txn.Description,
		amountStr,
		txn.Date.Format("2006-01-02"))

	warning := style.WarningStyle.Render("This action cannot be undone!")
	help := "[y] Yes, Delete • [n] Cancel"

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		warning,
		"",
		help,
	)

	return dialogStyle.Render(content)
}

// Load all invoices from all credit cards
func (m *TransactionsModel) loadAllInvoices() tea.Msg {
	var allInvoices []*entity.CreditCardInvoice
	
	for _, card := range m.creditCards {
		invoices, err := m.creditCardInvoiceUseCase.ListInvoicesByCard(m.ctx, card.ID)
		if err != nil {
			continue // Skip cards with errors
		}
		allInvoices = append(allInvoices, invoices...)
	}
	
	return invoicesLoadedMsg{invoices: allInvoices}
}

// Load transactions for a specific invoice
func (m *TransactionsModel) loadInvoiceTransactions(invoiceID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		transactions, err := m.transactionUseCase.GetTransactionsByCreditCardInvoice(m.ctx, invoiceID)
		if err != nil {
			return errMsg{err: err}
		}
		return invoiceTransactionsLoadedMsg{transactions: transactions}
	}
}

// Handle keys for invoice list view
func (m *TransactionsModel) handleInvoicesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "b":
		m.viewMode = TransactionViewList
	case "up", "k":
		if m.invoiceModel.selectedInvoiceIndex > 0 {
			m.invoiceModel.selectedInvoiceIndex--
		}
	case "down", "j":
		if m.invoiceModel.selectedInvoiceIndex < len(m.invoiceModel.invoices)-1 {
			m.invoiceModel.selectedInvoiceIndex++
		}
	case "enter":
		if m.invoiceModel.selectedInvoiceIndex < len(m.invoiceModel.invoices) {
			invoice := m.invoiceModel.invoices[m.invoiceModel.selectedInvoiceIndex]
			m.invoiceModel.selectedInvoice = invoice
			m.viewMode = TransactionViewInvoiceTransactions
			m.loading = true
			return m, m.loadInvoiceTransactions(invoice.ID)
		}
	}
	
	return m, nil
}

// Handle keys for invoice transactions view
func (m *TransactionsModel) handleInvoiceTransactionsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "b":
		m.viewMode = TransactionViewInvoices
		m.invoiceModel.invoiceTransactions = nil
		m.invoiceModel.selectedInvoice = nil
	case "up", "k":
		if m.invoiceModel.selectedTransactionIndex > 0 {
			m.invoiceModel.selectedTransactionIndex--
		}
	case "down", "j":
		if m.invoiceModel.selectedTransactionIndex < len(m.invoiceModel.invoiceTransactions)-1 {
			m.invoiceModel.selectedTransactionIndex++
		}
	}
	
	return m, nil
}

// Render invoice list
func (m *TransactionsModel) renderInvoicesList() string {
	var sections []string
	
	title := style.TitleStyle.Render("📋 Credit Card Invoices")
	sections = append(sections, title)
	
	if len(m.invoiceModel.invoices) == 0 {
		empty := style.InfoStyle.Render("No invoices found. Make sure you have credit cards with invoices.")
		sections = append(sections, empty)
	} else {
		table := m.renderInvoicesTable()
		sections = append(sections, table)
	}
	
	help := "[↑/↓] Navigate • [Enter] View Transactions • [b] Back"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render invoices table
func (m *TransactionsModel) renderInvoicesTable() string {
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)
	
	headers := []string{"Card", "Month", "Status", "Total", "Paid", "Balance", "Due Date"}
	headerRow := style.TableHeaderStyle.Render(
		fmt.Sprintf("%-15s %-8s %-8s %-12s %-12s %-12s %s",
			headers[0], headers[1], headers[2], headers[3], headers[4], headers[5], headers[6]),
	)
	
	var rows []string
	rows = append(rows, headerRow)
	
	for i, invoice := range m.invoiceModel.invoices {
		// Get card name
		cardName := "Unknown Card"
		for _, card := range m.creditCards {
			if card.ID == invoice.CreditCardID {
				cardName = card.Name
				break
			}
		}
		
		// Format amounts
		totalCharges := fmt.Sprintf("R$ %.2f", invoice.TotalCharges.Amount())
		paidAmount := fmt.Sprintf("R$ %.2f", invoice.TotalPayments.Amount())
		balanceAmount := fmt.Sprintf("R$ %.2f", invoice.ClosingBalance.Amount())
		
		// Format due date
		dueDate := invoice.DueDate.Format("2006-01-02")
		
		row := fmt.Sprintf("%-15s %-8s %-8s %-12s %-12s %-12s %s",
			cardName, invoice.ReferenceMonth, string(invoice.Status),
			totalCharges, paidAmount, balanceAmount, dueDate)
		
		if i == m.invoiceModel.selectedInvoiceIndex {
			row = style.SelectedMenuItemStyle.Render(row)
		} else {
			row = style.MenuItemStyle.Render(row)
		}
		
		rows = append(rows, row)
	}
	
	content := strings.Join(rows, "\n")
	return tableStyle.Render(content)
}

// Render invoice transactions
func (m *TransactionsModel) renderInvoiceTransactions() string {
	if m.invoiceModel.selectedInvoice == nil {
		return style.ErrorStyle.Render("No invoice selected")
	}
	
	var sections []string
	
	title := style.TitleStyle.Render(fmt.Sprintf("💳 Transactions - %s (%s)",
		m.getCardNameForInvoice(m.invoiceModel.selectedInvoice.CreditCardID),
		m.invoiceModel.selectedInvoice.ReferenceMonth))
	sections = append(sections, title)
	
	// Invoice summary
	summary := m.renderInvoiceSummary()
	sections = append(sections, summary)
	
	if len(m.invoiceModel.invoiceTransactions) == 0 {
		empty := style.InfoStyle.Render("No transactions found for this invoice.")
		sections = append(sections, empty)
	} else {
		table := m.renderInvoiceTransactionsTable()
		sections = append(sections, table)
	}
	
	help := "[↑/↓] Navigate • [b] Back to Invoices"
	sections = append(sections, style.HelpStyle.MarginTop(1).Render(help))
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// Render invoice summary
func (m *TransactionsModel) renderInvoiceSummary() string {
	invoice := m.invoiceModel.selectedInvoice
	
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Info).
		Padding(1, 2).
		MarginTop(1)
	
	var summary []string
	summary = append(summary, style.HeaderStyle.Render("Invoice Summary"))
	summary = append(summary, fmt.Sprintf("Period: %s to %s",
		invoice.OpeningDate.Format("2006-01-02"),
		invoice.ClosingDate.Format("2006-01-02")))
	summary = append(summary, fmt.Sprintf("Status: %s", string(invoice.Status)))
	summary = append(summary, fmt.Sprintf("Total Charges: R$ %.2f", invoice.TotalCharges.Amount()))
	summary = append(summary, fmt.Sprintf("Paid Amount: R$ %.2f", invoice.TotalPayments.Amount()))
	summary = append(summary, fmt.Sprintf("Balance: R$ %.2f", invoice.ClosingBalance.Amount()))
	summary = append(summary, fmt.Sprintf("Due Date: %s", invoice.DueDate.Format("2006-01-02")))
	
	content := strings.Join(summary, "\n")
	return summaryStyle.Render(content)
}

// Render invoice transactions table
func (m *TransactionsModel) renderInvoiceTransactionsTable() string {
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)
	
	headers := []string{"Date", "Description", "Category", "Amount"}
	headerRow := style.TableHeaderStyle.Render(
		fmt.Sprintf("%-12s %-30s %-15s %-12s",
			headers[0], headers[1], headers[2], headers[3]),
	)
	
	var rows []string
	rows = append(rows, headerRow)
	
	for i, txn := range m.invoiceModel.invoiceTransactions {
		amountStr := fmt.Sprintf("R$ %.2f", txn.Amount.Amount())
		if txn.Type == entity.TransactionTypeCredit {
			amountStr = style.SuccessStyle.Render("+" + amountStr)
		} else {
			amountStr = style.ErrorStyle.Render("-" + amountStr)
		}
		
		row := fmt.Sprintf("%-12s %-30s %-15s %-12s",
			txn.Date.Format("2006-01-02"),
			m.truncateString(txn.Description, 30),
			m.getCategoryDisplay(txn.Category),
			amountStr)
		
		if i == m.invoiceModel.selectedTransactionIndex {
			row = style.SelectedMenuItemStyle.Render(row)
		} else {
			row = style.MenuItemStyle.Render(row)
		}
		
		rows = append(rows, row)
	}
	
	content := strings.Join(rows, "\n")
	return tableStyle.Render(content)
}

// Helper to get card name for invoice
func (m *TransactionsModel) getCardNameForInvoice(cardID uuid.UUID) string {
	for _, card := range m.creditCards {
		if card.ID == cardID {
			return card.Name
		}
	}
	return "Unknown Card"
}

// Helper to truncate string
func (m *TransactionsModel) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
