package screen

import (
	"context"
	"financli/internal/application/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

// Placeholder implementations for other screens


func NewCreditCardsModel(ctx context.Context, creditCardUC *usecase.CreditCardUseCase, accountUC *usecase.AccountUseCase) tea.Model {
	return &simpleModel{title: "Credit Cards", content: "Credit card management screen - Coming soon!"}
}

func NewBillsModel(ctx context.Context, billUC *usecase.BillUseCase) tea.Model {
	return &simpleModel{title: "Bills", content: "Bill management screen - Coming soon!"}
}

func NewTransactionsModel(ctx context.Context, txnUC *usecase.TransactionUseCase, accountUC *usecase.AccountUseCase, cardUC *usecase.CreditCardUseCase, billUC *usecase.BillUseCase, personUC *usecase.PersonUseCase) tea.Model {
	return &simpleModel{title: "Transactions", content: "Transaction management screen - Coming soon!"}
}

func NewPeopleModel(ctx context.Context, personUC *usecase.PersonUseCase) tea.Model {
	return &simpleModel{title: "People", content: "People management screen - Coming soon!"}
}

func NewReportsModel(ctx context.Context, reportUC *usecase.ReportUseCase, personUC *usecase.PersonUseCase, billUC *usecase.BillUseCase) tea.Model {
	return &simpleModel{title: "Reports", content: "Reports screen - Coming soon!"}
}

type simpleModel struct {
	title   string
	content string
}

func (m *simpleModel) Init() tea.Cmd {
	return nil
}

func (m *simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *simpleModel) View() string {
	return m.content
}