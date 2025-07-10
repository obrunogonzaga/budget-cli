package screen

import (
	"context"
	"fmt"
	"strings"

	"financli/internal/application/usecase"
	"financli/internal/domain/entity"
	"financli/internal/interfaces/tui/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type PeopleModel struct {
	ctx           context.Context
	personUseCase *usecase.PersonUseCase

	people        []*entity.Person
	selectedIndex int
	viewMode      PeopleViewMode

	loading bool
	err     error

	// Form state
	formModel         *PersonFormModel
	showConfirmDelete bool

	width  int
	height int
}

type PeopleViewMode int

const (
	PeopleViewList PeopleViewMode = iota
	PeopleViewForm
	PeopleViewConfirm
)

type PersonFormModel struct {
	name  string
	email string
	phone string

	focusedField int
	editing      bool
	editingID    *uuid.UUID

	nameInput  string
	emailInput string
	phoneInput string
}

func NewPeopleModel(ctx context.Context, personUC *usecase.PersonUseCase) tea.Model {
	return &PeopleModel{
		ctx:           ctx,
		personUseCase: personUC,
		viewMode:      PeopleViewList,
		loading:       true,
		formModel:     &PersonFormModel{},
	}
}

func (m *PeopleModel) Init() tea.Cmd {
	return m.loadPeople
}

func (m *PeopleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case peopleLoadedMsg:
		m.loading = false
		m.people = msg.people
		if len(m.people) > 0 && m.selectedIndex >= len(m.people) {
			m.selectedIndex = len(m.people) - 1
		}
		return m, nil

	case personActionMsg:
		m.loading = false
		m.viewMode = PeopleViewList
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
		return m, m.loadPeople

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch m.viewMode {
		case PeopleViewList:
			return m.handleListKeys(msg)
		case PeopleViewForm:
			return m.handleFormKeys(msg)
		case PeopleViewConfirm:
			return m.handleConfirmKeys(msg)
		}
	}

	return m, nil
}

func (m *PeopleModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.people)-1 {
			m.selectedIndex++
		}
	case "enter":
		if len(m.people) > 0 {
			return m.showPersonDetails()
		}
	case "n":
		m.viewMode = PeopleViewForm
		m.formModel.editing = false
		m.formModel.editingID = nil
		m.resetForm()
	case "e":
		if len(m.people) > 0 {
			return m.editPerson()
		}
	case "d":
		if len(m.people) > 0 {
			m.viewMode = PeopleViewConfirm
			m.showConfirmDelete = true
		}
	case "r":
		m.loading = true
		return m, m.loadPeople
	case "b":
		// Go back to dashboard
		return m, func() tea.Msg { return BackToDashboardMsg{} }
	}

	return m, nil
}

func (m *PeopleModel) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = PeopleViewList
		m.resetForm()
	case "tab", "down":
		m.formModel.focusedField = (m.formModel.focusedField + 1) % 5
	case "shift+tab", "up":
		m.formModel.focusedField = (m.formModel.focusedField - 1 + 5) % 5
	case "enter":
		if m.formModel.focusedField == 3 {
			return m.submitForm()
		} else if m.formModel.focusedField == 4 {
			// Cancel button
			m.viewMode = PeopleViewList
			m.resetForm()
		}
	default:
		return m.handleFormInput(msg)
	}

	return m, nil
}

func (m *PeopleModel) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only handle input for fields 0-2 (name, email, phone)
	// Fields 3-4 are buttons
	if m.formModel.focusedField > 2 {
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
		case "backspace":
			if len(m.formModel.emailInput) > 0 {
				m.formModel.emailInput = m.formModel.emailInput[:len(m.formModel.emailInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.formModel.emailInput += msg.String()
			}
		}
	case 2:
		switch msg.String() {
		case "backspace":
			if len(m.formModel.phoneInput) > 0 {
				m.formModel.phoneInput = m.formModel.phoneInput[:len(m.formModel.phoneInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.formModel.phoneInput += msg.String()
			}
		}
	}

	return m, nil
}

func (m *PeopleModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		return m.deletePerson()
	case "n", "esc":
		m.viewMode = PeopleViewList
		m.showConfirmDelete = false
	}

	return m, nil
}

func (m *PeopleModel) showPersonDetails() (tea.Model, tea.Cmd) {
	// For now, just switch to edit mode
	return m.editPerson()
}

func (m *PeopleModel) editPerson() (tea.Model, tea.Cmd) {
	if len(m.people) == 0 {
		return m, nil
	}

	person := m.people[m.selectedIndex]
	m.viewMode = PeopleViewForm
	m.formModel.editing = true
	m.formModel.editingID = &person.ID
	m.formModel.nameInput = person.Name
	m.formModel.emailInput = person.Email
	m.formModel.phoneInput = person.Phone
	m.formModel.focusedField = 0

	return m, nil
}

func (m *PeopleModel) submitForm() (tea.Model, tea.Cmd) {
	// Validate name is required
	if strings.TrimSpace(m.formModel.nameInput) == "" {
		m.err = fmt.Errorf("name is required")
		return m, nil
	}

	m.loading = true
	m.err = nil

	if m.formModel.editing && m.formModel.editingID != nil {
		return m, m.updatePerson
	} else {
		return m, m.createPerson
	}
}

func (m *PeopleModel) deletePerson() (tea.Model, tea.Cmd) {
	if len(m.people) == 0 {
		return m, nil
	}

	m.loading = true
	m.showConfirmDelete = false
	return m, m.deletePersonCmd
}

func (m *PeopleModel) resetForm() {
	m.formModel.nameInput = ""
	m.formModel.emailInput = ""
	m.formModel.phoneInput = ""
	m.formModel.focusedField = 0
	m.formModel.editing = false
	m.formModel.editingID = nil
}

func (m *PeopleModel) View() string {
	if m.loading {
		return style.InfoStyle.Render("Loading people...")
	}

	switch m.viewMode {
	case PeopleViewList:
		return m.renderList()
	case PeopleViewForm:
		return m.renderForm()
	case PeopleViewConfirm:
		return m.renderConfirm()
	}

	return ""
}

func (m *PeopleModel) renderList() string {
	var content strings.Builder

	content.WriteString(style.HeaderStyle.Render("People Management"))
	content.WriteString("\n\n")

	if m.err != nil {
		content.WriteString(style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		content.WriteString("\n\n")
	}

	if len(m.people) == 0 {
		content.WriteString(style.InfoStyle.Render("No people registered yet."))
		content.WriteString("\n\n")
		content.WriteString("Press 'n' to add a new person.")
	} else {
		// Table header
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(style.Primary).
			PaddingLeft(2).
			PaddingRight(2)

		content.WriteString(headerStyle.Render("Name"))
		content.WriteString(headerStyle.Width(30).Render("Email"))
		content.WriteString(headerStyle.Width(20).Render("Phone"))
		content.WriteString(headerStyle.Width(12).Render("Created"))
		content.WriteString("\n")

		// Table rows
		for i, person := range m.people {
			var rowStyle lipgloss.Style
			if i == m.selectedIndex {
				rowStyle = style.SelectedMenuItemStyle
			} else {
				rowStyle = style.MenuItemStyle
			}

			// Format created date
			createdDate := person.CreatedAt.Format("2006-01-02")

			row := fmt.Sprintf("%-25s %-30s %-20s %s",
				person.Name,
				person.Email,
				person.Phone,
				createdDate,
			)

			content.WriteString(rowStyle.Render(row))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(style.HelpStyle.Render("[n] New • [e] Edit • [d] Delete • [r] Refresh • [b] Back • [q] Quit"))

	return content.String()
}

func (m *PeopleModel) renderForm() string {
	var content strings.Builder

	if m.formModel.editing {
		content.WriteString(style.HeaderStyle.Render("Edit Person"))
	} else {
		content.WriteString(style.HeaderStyle.Render("Add New Person"))
	}
	content.WriteString("\n\n")

	if m.err != nil {
		content.WriteString(style.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		content.WriteString("\n\n")
	}

	// Name field
	nameStyle := style.InputStyle
	if m.formModel.focusedField == 0 {
		nameStyle = style.FocusedInputStyle
	}
	content.WriteString(style.HeaderStyle.Render("Name (required):"))
	content.WriteString("\n")
	content.WriteString(nameStyle.Render(m.formModel.nameInput))
	content.WriteString("\n\n")

	// Email field
	emailStyle := style.InputStyle
	if m.formModel.focusedField == 1 {
		emailStyle = style.FocusedInputStyle
	}
	content.WriteString(style.HeaderStyle.Render("Email:"))
	content.WriteString("\n")
	content.WriteString(emailStyle.Render(m.formModel.emailInput))
	content.WriteString("\n\n")

	// Phone field
	phoneStyle := style.InputStyle
	if m.formModel.focusedField == 2 {
		phoneStyle = style.FocusedInputStyle
	}
	content.WriteString(style.HeaderStyle.Render("Phone:"))
	content.WriteString("\n")
	content.WriteString(phoneStyle.Render(m.formModel.phoneInput))
	content.WriteString("\n\n")

	// Submit button
	submitStyle := style.ButtonStyle
	if m.formModel.focusedField == 3 {
		submitStyle = style.ButtonStyle.Background(style.Success)
	}
	submitText := "Create"
	if m.formModel.editing {
		submitText = "Update"
	}
	content.WriteString(submitStyle.Render(submitText))
	content.WriteString("  ")

	// Cancel button
	cancelStyle := style.SecondaryButtonStyle
	if m.formModel.focusedField == 4 {
		cancelStyle = style.ButtonStyle.Background(style.Danger)
	}
	content.WriteString(cancelStyle.Render("Cancel"))
	content.WriteString("\n\n")

	content.WriteString(style.HelpStyle.Render("[Tab] Next • [Shift+Tab] Previous • [Enter] Submit • [Esc] Cancel"))

	return content.String()
}

func (m *PeopleModel) renderConfirm() string {
	var content strings.Builder

	content.WriteString(style.HeaderStyle.Render("Confirm Deletion"))
	content.WriteString("\n\n")

	if len(m.people) > 0 {
		person := m.people[m.selectedIndex]
		content.WriteString(fmt.Sprintf("Are you sure you want to delete '%s'?", person.Name))
		content.WriteString("\n\n")
		content.WriteString(style.WarningStyle.Render("This action cannot be undone."))
		content.WriteString("\n\n")
	}

	content.WriteString(style.HelpStyle.Render("[y] Yes • [n] No"))

	return content.String()
}

// IsInFormMode implements the FormModeChecker interface
func (m *PeopleModel) IsInFormMode() bool {
	return m.viewMode == PeopleViewForm || m.viewMode == PeopleViewConfirm
}

// Message types
type personActionMsg struct{}

// Commands
func (m *PeopleModel) loadPeople() tea.Msg {
	people, err := m.personUseCase.ListPeople(m.ctx)
	if err != nil {
		return errMsg{err: err}
	}
	return peopleLoadedMsg{people: people}
}

func (m *PeopleModel) createPerson() tea.Msg {
	_, err := m.personUseCase.CreatePerson(
		m.ctx,
		strings.TrimSpace(m.formModel.nameInput),
		strings.TrimSpace(m.formModel.emailInput),
		strings.TrimSpace(m.formModel.phoneInput),
	)
	if err != nil {
		return errMsg{err: err}
	}
	return personActionMsg{}
}

func (m *PeopleModel) updatePerson() tea.Msg {
	if m.formModel.editingID == nil {
		return errMsg{err: fmt.Errorf("no person selected for editing")}
	}

	err := m.personUseCase.UpdatePerson(
		m.ctx,
		*m.formModel.editingID,
		strings.TrimSpace(m.formModel.nameInput),
		strings.TrimSpace(m.formModel.emailInput),
		strings.TrimSpace(m.formModel.phoneInput),
	)
	if err != nil {
		return errMsg{err: err}
	}
	return personActionMsg{}
}

func (m *PeopleModel) deletePersonCmd() tea.Msg {
	if len(m.people) == 0 {
		return errMsg{err: fmt.Errorf("no person selected for deletion")}
	}

	person := m.people[m.selectedIndex]
	err := m.personUseCase.DeletePerson(m.ctx, person.ID)
	if err != nil {
		return errMsg{err: err}
	}

	return personActionMsg{}
}