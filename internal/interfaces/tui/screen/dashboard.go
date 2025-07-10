package screen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"financli/internal/application/usecase"
	"financli/internal/domain/entity"
	"financli/internal/interfaces/tui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

type DashboardModel struct {
	ctx                context.Context
	accountUseCase     *usecase.AccountUseCase
	transactionUseCase *usecase.TransactionUseCase
	billUseCase        *usecase.BillUseCase

	accounts     []*entity.Account
	recentTxns   []*entity.Transaction
	pendingBills []*entity.Bill

	totalBalance    float64
	monthlyIncome   float64
	monthlyExpenses float64

	loading bool
	err     error
}

func NewDashboardModel(ctx context.Context, accountUC *usecase.AccountUseCase, txnUC *usecase.TransactionUseCase, billUC *usecase.BillUseCase) tea.Model {
	return &DashboardModel{
		ctx:                ctx,
		accountUseCase:     accountUC,
		transactionUseCase: txnUC,
		billUseCase:        billUC,
		loading:            true,
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return m.loadData
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		m.loading = false
		m.accounts = msg.accounts
		m.recentTxns = msg.transactions
		m.pendingBills = msg.bills
		m.calculateTotals()
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m *DashboardModel) View() string {
	if m.loading {
		return style.InfoStyle.Render("Loading dashboard data...")
	}

	if m.err != nil {
		return style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var sections []string

	// Summary Cards
	summaryCards := m.renderSummaryCards()
	sections = append(sections, summaryCards)

	// Monthly Trend Chart
	trendChart := m.renderMonthlyTrend()
	sections = append(sections, trendChart)

	// Bottom Section: Accounts, Recent Transactions, Pending Bills
	bottomSection := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderAccountsList(),
		m.renderRecentTransactions(),
		m.renderPendingBills(),
	)
	sections = append(sections, bottomSection)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m *DashboardModel) renderSummaryCards() string {
	cards := []string{
		m.renderCard("Total Balance", fmt.Sprintf("R$ %.2f", m.totalBalance), style.Primary),
		m.renderCard("Monthly Income", fmt.Sprintf("R$ %.2f", m.monthlyIncome), style.Success),
		m.renderCard("Monthly Expenses", fmt.Sprintf("R$ %.2f", m.monthlyExpenses), style.Danger),
		m.renderCard("Net Savings", fmt.Sprintf("R$ %.2f", m.monthlyIncome-m.monthlyExpenses), style.Info),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func (m *DashboardModel) renderCard(title, value string, color lipgloss.Color) string {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 2).
		Width(20).
		Height(5)

	titleStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(style.Text).
		Bold(true).
		MarginTop(1)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		titleStyle.Render(title),
		valueStyle.Render(value),
	)

	return cardStyle.Render(content)
}

func (m *DashboardModel) renderMonthlyTrend() string {
	// Generate sample data for the last 30 days
	data := make([]float64, 30)
	for i := range data {
		// Simulate daily balance changes
		data[i] = m.totalBalance + float64(i-15)*100 + float64(i%7)*50
	}

	graph := asciigraph.Plot(data,
		asciigraph.Height(10),
		asciigraph.Width(80),
		asciigraph.Caption("30-Day Balance Trend"),
	)

	chartStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		MarginTop(1)

	return chartStyle.Render(graph)
}

func (m *DashboardModel) renderAccountsList() string {
	title := style.TitleStyle.Render("Accounts")

	if len(m.accounts) == 0 {
		return m.renderSection(title, "No accounts found", 30)
	}

	var lines []string
	for _, acc := range m.accounts {
		icon := m.getAccountIcon(acc.Type)
		line := fmt.Sprintf("%s %-15s %10s",
			icon,
			truncate(acc.Name, 15),
			acc.Balance.String(),
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return m.renderSection(title, content, 35)
}

func (m *DashboardModel) renderRecentTransactions() string {
	title := style.TitleStyle.Render("Recent Transactions")

	if len(m.recentTxns) == 0 {
		return m.renderSection(title, "No recent transactions", 40)
	}

	var lines []string
	for i, txn := range m.recentTxns {
		if i >= 5 {
			break
		}

		icon := "📤"
		if txn.Type == entity.TransactionTypeCredit {
			icon = "📥"
		}

		line := fmt.Sprintf("%s %s %-15s %10s",
			icon,
			txn.Date.Format("Jan 02"),
			truncate(txn.Description, 15),
			txn.Amount.String(),
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return m.renderSection(title, content, 45)
}

func (m *DashboardModel) renderPendingBills() string {
	title := style.TitleStyle.Render("Pending Bills")

	if len(m.pendingBills) == 0 {
		return m.renderSection(title, "No pending bills", 30)
	}

	var lines []string
	for i, bill := range m.pendingBills {
		if i >= 5 {
			break
		}

		remaining, _ := bill.GetRemainingAmount()
		statusIcon := m.getBillStatusIcon(bill.Status)

		line := fmt.Sprintf("%s %-15s %10s",
			statusIcon,
			truncate(bill.Name, 15),
			remaining.String(),
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return m.renderSection(title, content, 35)
}

func (m *DashboardModel) renderSection(title, content string, width int) string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.Border).
		Padding(1, 2).
		Width(width).
		Height(10).
		MarginRight(1)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		content,
	)

	return sectionStyle.Render(fullContent)
}

func (m *DashboardModel) calculateTotals() {
	m.totalBalance = 0
	for _, acc := range m.accounts {
		m.totalBalance += acc.Balance.Amount()
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	m.monthlyIncome = 0
	m.monthlyExpenses = 0

	for _, txn := range m.recentTxns {
		if txn.Date.After(startOfMonth) {
			if txn.Type == entity.TransactionTypeCredit {
				m.monthlyIncome += txn.Amount.Amount()
			} else {
				m.monthlyExpenses += txn.Amount.Amount()
			}
		}
	}
}

func (m *DashboardModel) getAccountIcon(accountType entity.AccountType) string {
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

func (m *DashboardModel) getBillStatusIcon(status entity.BillStatus) string {
	switch status {
	case entity.BillStatusOpen:
		return "📄"
	case entity.BillStatusPaid:
		return "✅"
	case entity.BillStatusOverdue:
		return "🔴"
	case entity.BillStatusClosed:
		return "📁"
	default:
		return "📋"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Commands
func (m *DashboardModel) loadData() tea.Msg {
	accounts, err := m.accountUseCase.ListAccounts(m.ctx)
	if err != nil {
		return errMsg{err: err}
	}

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	transactions, err := m.transactionUseCase.GetTransactionsByDateRange(m.ctx, thirtyDaysAgo, now)
	if err != nil {
		return errMsg{err: err}
	}

	bills, err := m.billUseCase.GetPendingBills(m.ctx)
	if err != nil {
		return errMsg{err: err}
	}

	return dataLoadedMsg{
		accounts:     accounts,
		transactions: transactions,
		bills:        bills,
	}
}

// Messages
type dataLoadedMsg struct {
	accounts     []*entity.Account
	transactions []*entity.Transaction
	bills        []*entity.Bill
}

type errMsg struct {
	err error
}
