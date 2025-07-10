package screen

import (
	"context"
	"financli/internal/application/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

// Placeholder implementations for other screens

// NewCreditCardsModel is now implemented in credit_cards.go

// NewBillsModel is now implemented in bills.go

// NewTransactionsModel is now implemented in transactions.go

// NewPeopleModel is now implemented in people.go

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
