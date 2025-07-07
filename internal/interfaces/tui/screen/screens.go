package screen

import (
	"context"
	"financli/internal/application/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

// Placeholder implementations for other screens


// NewCreditCardsModel is now implemented in credit_cards.go

func NewBillsModel(ctx context.Context, billUC *usecase.BillUseCase) tea.Model {
	return &simpleModel{title: "Bills", content: "Bill management screen - Coming soon!"}
}

// NewTransactionsModel is now implemented in transactions.go

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