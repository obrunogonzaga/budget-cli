package tui

import (
	"context"
	"fmt"

	"financli/internal/application/usecase"
	"financli/internal/interfaces/tui/screen"
	"financli/internal/interfaces/tui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormModeChecker interface for screens that can be in form mode
type FormModeChecker interface {
	IsInFormMode() bool
}

// Message types for inter-screen communication

type Screen int

const (
	DashboardScreen Screen = iota
	AccountsScreen
	CreditCardsScreen
	BillsScreen
	TransactionsScreen
	PeopleScreen
	ReportsScreen
)

type App struct {
	currentScreen     Screen
	dashboardModel    tea.Model
	accountsModel     tea.Model
	creditCardsModel  tea.Model
	billsModel        tea.Model
	transactionsModel tea.Model
	peopleModel       tea.Model
	reportsModel      tea.Model
	width             int
	height            int
	ctx               context.Context
}

type UseCases struct {
	Account           *usecase.AccountUseCase
	CreditCard        *usecase.CreditCardUseCase
	CreditCardInvoice *usecase.CreditCardInvoiceUseCase
	Bill              *usecase.BillUseCase
	Transaction       *usecase.TransactionUseCase
	Person            *usecase.PersonUseCase
	Report            *usecase.ReportUseCase
}

func NewApp(ctx context.Context, useCases UseCases) *App {
	return &App{
		currentScreen:     DashboardScreen,
		dashboardModel:    screen.NewDashboardModel(ctx, useCases.Account, useCases.Transaction, useCases.Bill),
		accountsModel:     screen.NewAccountsModel(ctx, useCases.Account),
		creditCardsModel:  screen.NewCreditCardsModel(ctx, useCases.CreditCard, useCases.CreditCardInvoice, useCases.Account),
		billsModel:        screen.NewBillsModel(ctx, useCases.Bill),
		transactionsModel: screen.NewTransactionsModel(ctx, useCases.Transaction, useCases.Account, useCases.CreditCard, useCases.CreditCardInvoice, useCases.Bill, useCases.Person),
		peopleModel:       screen.NewPeopleModel(ctx, useCases.Person),
		reportsModel:      screen.NewReportsModel(ctx, useCases.Report, useCases.Person, useCases.Bill),
		ctx:               ctx,
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.dashboardModel.Init(),
		tea.EnterAltScreen,
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if current screen is in form mode before handling navigation
		var isInFormMode bool
		switch a.currentScreen {
		case AccountsScreen:
			if checker, ok := a.accountsModel.(FormModeChecker); ok {
				isInFormMode = checker.IsInFormMode()
			}
		case TransactionsScreen:
			if checker, ok := a.transactionsModel.(FormModeChecker); ok {
				isInFormMode = checker.IsInFormMode()
			}
		case CreditCardsScreen:
			if checker, ok := a.creditCardsModel.(FormModeChecker); ok {
				isInFormMode = checker.IsInFormMode()
			}
		case BillsScreen:
			if checker, ok := a.billsModel.(FormModeChecker); ok {
				isInFormMode = checker.IsInFormMode()
			}
			// Add other screens here when they implement forms
		}

		// Only handle menu navigation if not in form mode
		if !isInFormMode {
			switch msg.String() {
			case "ctrl+c", "q":
				return a, tea.Quit
			case "1":
				a.currentScreen = DashboardScreen
				return a, a.dashboardModel.Init()
			case "2":
				a.currentScreen = AccountsScreen
				return a, a.accountsModel.Init()
			case "3":
				a.currentScreen = CreditCardsScreen
				return a, a.creditCardsModel.Init()
			case "4":
				a.currentScreen = BillsScreen
				return a, a.billsModel.Init()
			case "5":
				a.currentScreen = TransactionsScreen
				return a, a.transactionsModel.Init()
			case "6":
				a.currentScreen = PeopleScreen
				return a, a.peopleModel.Init()
			case "7":
				a.currentScreen = ReportsScreen
				return a, a.reportsModel.Init()
			}
		} else {
			// Always allow quit even in form mode
			switch msg.String() {
			case "ctrl+c":
				return a, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	default:
		// Handle custom messages from screens
		switch msg.(type) {
		case screen.BackToDashboardMsg:
			a.currentScreen = DashboardScreen
			return a, a.dashboardModel.Init()
		}
	}

	var cmd tea.Cmd
	switch a.currentScreen {
	case DashboardScreen:
		a.dashboardModel, cmd = a.dashboardModel.Update(msg)
	case AccountsScreen:
		a.accountsModel, cmd = a.accountsModel.Update(msg)
	case CreditCardsScreen:
		a.creditCardsModel, cmd = a.creditCardsModel.Update(msg)
	case BillsScreen:
		a.billsModel, cmd = a.billsModel.Update(msg)
	case TransactionsScreen:
		a.transactionsModel, cmd = a.transactionsModel.Update(msg)
	case PeopleScreen:
		a.peopleModel, cmd = a.peopleModel.Update(msg)
	case ReportsScreen:
		a.reportsModel, cmd = a.reportsModel.Update(msg)
	}

	return a, cmd
}

func (a *App) View() string {
	header := a.renderHeader()

	var content string
	switch a.currentScreen {
	case DashboardScreen:
		content = a.dashboardModel.View()
	case AccountsScreen:
		content = a.accountsModel.View()
	case CreditCardsScreen:
		content = a.creditCardsModel.View()
	case BillsScreen:
		content = a.billsModel.View()
	case TransactionsScreen:
		content = a.transactionsModel.View()
	case PeopleScreen:
		content = a.peopleModel.View()
	case ReportsScreen:
		content = a.reportsModel.View()
	}

	help := a.renderHelp()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		help,
	)
}

func (a *App) renderHeader() string {
	menu := []string{
		"[1] Dashboard",
		"[2] Accounts",
		"[3] Credit Cards",
		"[4] Bills",
		"[5] Transactions",
		"[6] People",
		"[7] Reports",
	}

	for i, item := range menu {
		if Screen(i) == a.currentScreen {
			menu[i] = style.SelectedMenuItemStyle.Render(fmt.Sprintf("● %s", item[4:]))
		} else {
			menu[i] = style.MenuItemStyle.Render(item)
		}
	}

	menuBar := lipgloss.JoinHorizontal(lipgloss.Top, menu...)

	title := style.TitleStyle.Render("💰 FinanCLI - Personal Finance Manager")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		menuBar,
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(style.Border).
			Width(a.width).
			Render(""),
	)
}

func (a *App) renderHelp() string {
	help := "[q] Quit • [1-7] Navigate • [↑/↓] Select • [Enter] Confirm • [Esc] Cancel"
	return style.HelpStyle.
		Width(a.width).
		Align(lipgloss.Center).
		MarginTop(1).
		Render(help)
}
