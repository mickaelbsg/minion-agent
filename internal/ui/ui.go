package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"minion/internal/admin"
	"minion/internal/storage"
)

type screen int

const (
	menuScreen screen = iota
	statusScreen
	setupFormScreen
	setupConfirmScreen
	configFormScreen
	configConfirmScreen
	clientsListScreen
	clientCreateFormScreen
	clientCreateConfirmScreen
	clientActionConfirmScreen
	messageScreen
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	focusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)

type Model struct {
	service     *admin.Service
	current     screen
	returnTo    screen
	menuItems   []string
	menuIndex   int
	clientIndex int
	inputs      []textinput.Model
	inputIndex  int
	status      admin.Status
	clients     []storage.Client
	setupDraft  admin.SetupOptions
	configDraft admin.ConfigUpdate
	clientDraft struct {
		name string
		ips  string
	}
	clientAction struct {
		name    string
		enabled bool
		delete  bool
	}
	message   string
	messageOK bool
}

func Run(configPath, section string) error {
	if err := ensureInteractive(os.Stdin, os.Stdout); err != nil {
		return err
	}

	service := admin.NewService(configPath)
	model := NewModel(service, section)
	prog := tea.NewProgram(model)
	if _, err := prog.Run(); err != nil {
		return err
	}
	return nil
}

func NewModel(service *admin.Service, section string) Model {
	model := Model{
		service: service,
		current: menuScreen,
		menuItems: []string{
			"Setup inicial",
			"Configuracao do servico",
			"Clientes",
			"Status resumido",
			"Sair",
		},
	}

	switch section {
	case "", "menu":
	case "setup":
		model.enterSetup()
	case "config":
		model.enterConfig()
	case "clients":
		model.enterClients()
	case "status":
		model.enterStatus()
	default:
		model.showMessage(
			fmt.Sprintf("Secao invalida %q. Use setup, config, clients ou status.", section),
			false,
			menuScreen,
		)
	}

	return model
}

func ensureInteractive(stdin, stdout *os.File) error {
	if !term.IsTerminal(int(stdin.Fd())) || !term.IsTerminal(int(stdout.Fd())) {
		return fmt.Errorf("interactive UI requires a TTY; use `minion setup` or `minion client ...` for non-interactive usage")
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key := msg.String(); key == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.current {
		case menuScreen:
			return m.updateMenu(msg)
		case statusScreen:
			return m.updateStatus(msg)
		case setupFormScreen:
			return m.updateSetupForm(msg)
		case setupConfirmScreen:
			return m.updateSetupConfirm(msg)
		case configFormScreen:
			return m.updateConfigForm(msg)
		case configConfirmScreen:
			return m.updateConfigConfirm(msg)
		case clientsListScreen:
			return m.updateClients(msg)
		case clientCreateFormScreen:
			return m.updateClientCreateForm(msg)
		case clientCreateConfirmScreen:
			return m.updateClientCreateConfirm(msg)
		case clientActionConfirmScreen:
			return m.updateClientActionConfirm(msg)
		case messageScreen:
			return m.updateMessage(msg)
		}
	}

	return m, nil
}

func (m Model) View() string {
	switch m.current {
	case menuScreen:
		return m.viewMenu()
	case statusScreen:
		return m.viewStatus()
	case setupFormScreen:
		return m.viewSetupForm()
	case setupConfirmScreen:
		return m.viewSetupConfirm()
	case configFormScreen:
		return m.viewConfigForm()
	case configConfirmScreen:
		return m.viewConfigConfirm()
	case clientsListScreen:
		return m.viewClients()
	case clientCreateFormScreen:
		return m.viewClientCreateForm()
	case clientCreateConfirmScreen:
		return m.viewClientCreateConfirm()
	case clientActionConfirmScreen:
		return m.viewClientActionConfirm()
	case messageScreen:
		return m.viewMessage()
	default:
		return ""
	}
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
	case "down", "j":
		if m.menuIndex < len(m.menuItems)-1 {
			m.menuIndex++
		}
	case "enter":
		switch m.menuIndex {
		case 0:
			m.enterSetup()
		case 1:
			m.enterConfig()
		case 2:
			m.enterClients()
		case 3:
			m.enterStatus()
		case 4:
			return m, tea.Quit
		}
	}
	return m, textinput.Blink
}

func (m Model) updateStatus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.enterStatus()
	case "esc", "backspace":
		m.current = menuScreen
	}
	return m, nil
}

func (m Model) updateSetupForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = menuScreen
		return m, nil
	case "enter":
		if m.inputIndex < len(m.inputs)-1 {
			m.inputIndex++
			m.focusInput(m.inputIndex)
			return m, nil
		}

		m.setupDraft = admin.SetupOptions{
			ClientName: strings.TrimSpace(m.inputs[0].Value()),
			ClientIPs:  strings.TrimSpace(m.inputs[1].Value()),
		}
		m.current = setupConfirmScreen
		return m, nil
	case "up", "shift+tab":
		if m.inputIndex > 0 {
			m.inputIndex--
			m.focusInput(m.inputIndex)
		}
		return m, nil
	case "down", "tab":
		if m.inputIndex < len(m.inputs)-1 {
			m.inputIndex++
			m.focusInput(m.inputIndex)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg)
	return m, cmd
}

func (m Model) updateSetupConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = setupFormScreen
	case "enter":
		result, err := m.service.Setup(m.setupDraft)
		if err != nil {
			m.showMessage(err.Error(), false, menuScreen)
			return m, nil
		}

		lines := []string{
			"Setup concluido.",
			fmt.Sprintf("Config: %s", result.ConfigPath),
			fmt.Sprintf("DB: %s", result.DBPath),
			fmt.Sprintf("TLS cert: %s", result.TLSCertPath),
		}
		if result.BootstrapCreated {
			lines = append(lines,
				fmt.Sprintf("Cliente bootstrap: %s", result.ClientName),
				fmt.Sprintf("IPs permitidos: %s", result.ClientIPs),
				fmt.Sprintf("API key: %s", result.APIKey),
			)
		} else {
			lines = append(lines, "Clientes existentes encontrados; nenhuma nova API key foi gerada.")
		}
		m.showMessage(strings.Join(lines, "\n"), true, menuScreen)
	}
	return m, nil
}

func (m Model) updateConfigForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = menuScreen
		return m, nil
	case "enter":
		if m.inputIndex < len(m.inputs)-1 {
			m.inputIndex++
			m.focusInput(m.inputIndex)
			return m, nil
		}

		m.configDraft.Bind = strings.TrimSpace(m.inputs[0].Value())
		m.configDraft.DBPath = strings.TrimSpace(m.inputs[1].Value())
		allowInsecure := strings.TrimSpace(strings.ToLower(m.inputs[2].Value()))
		m.configDraft.AllowInsecureHTTP = allowInsecure == "true" || allowInsecure == "1" || allowInsecure == "yes"
		m.current = configConfirmScreen
		return m, nil
	case "up", "shift+tab":
		if m.inputIndex > 0 {
			m.inputIndex--
			m.focusInput(m.inputIndex)
		}
		return m, nil
	case "down", "tab":
		if m.inputIndex < len(m.inputs)-1 {
			m.inputIndex++
			m.focusInput(m.inputIndex)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg)
	return m, cmd
}

func (m Model) updateConfigConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = configFormScreen
	case "enter":
		if err := m.service.SaveConfig(m.configDraft); err != nil {
			m.showMessage(err.Error(), false, menuScreen)
			return m, nil
		}
		m.showMessage("Configuracao salva com sucesso.", true, menuScreen)
	}
	return m, nil
}

func (m Model) updateClients(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = menuScreen
	case "r":
		m.enterClients()
	case "up", "k":
		if m.clientIndex > 0 {
			m.clientIndex--
		}
	case "down", "j":
		if m.clientIndex < len(m.clients)-1 {
			m.clientIndex++
		}
	case "a":
		m.enterClientCreate()
	case "e":
		if len(m.clients) == 0 {
			return m, nil
		}
		client := m.clients[m.clientIndex]
		m.clientAction = struct {
			name    string
			enabled bool
			delete  bool
		}{name: client.Name, enabled: !client.Enabled}
		m.current = clientActionConfirmScreen
	case "d":
		if len(m.clients) == 0 {
			return m, nil
		}
		client := m.clients[m.clientIndex]
		m.clientAction = struct {
			name    string
			enabled bool
			delete  bool
		}{name: client.Name, delete: true}
		m.current = clientActionConfirmScreen
	}
	return m, nil
}

func (m Model) updateClientCreateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = clientsListScreen
		return m, nil
	case "enter":
		if m.inputIndex < len(m.inputs)-1 {
			m.inputIndex++
			m.focusInput(m.inputIndex)
			return m, nil
		}
		m.clientDraft.name = strings.TrimSpace(m.inputs[0].Value())
		m.clientDraft.ips = strings.TrimSpace(m.inputs[1].Value())
		m.current = clientCreateConfirmScreen
		return m, nil
	case "up", "shift+tab":
		if m.inputIndex > 0 {
			m.inputIndex--
			m.focusInput(m.inputIndex)
		}
		return m, nil
	case "down", "tab":
		if m.inputIndex < len(m.inputs)-1 {
			m.inputIndex++
			m.focusInput(m.inputIndex)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.inputIndex], cmd = m.inputs[m.inputIndex].Update(msg)
	return m, cmd
}

func (m Model) updateClientCreateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = clientCreateFormScreen
	case "enter":
		client, err := m.service.CreateClient(m.clientDraft.name, m.clientDraft.ips)
		if err != nil {
			m.showMessage(err.Error(), false, clientsListScreen)
			return m, nil
		}
		m.showMessage(
			fmt.Sprintf("Cliente criado.\nNome: %s\nIPs: %s\nAPI key: %s", client.Name, client.AllowedIPs, client.APIKey),
			true,
			clientsListScreen,
		)
	}
	return m, nil
}

func (m Model) updateClientActionConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = clientsListScreen
	case "enter":
		var err error
		if m.clientAction.delete {
			err = m.service.DeleteClient(m.clientAction.name)
		} else {
			err = m.service.SetClientEnabled(m.clientAction.name, m.clientAction.enabled)
		}
		if err != nil {
			m.showMessage(err.Error(), false, clientsListScreen)
			return m, nil
		}
		m.enterClients()
		action := "habilitado"
		if m.clientAction.delete {
			action = "removido"
		} else if !m.clientAction.enabled {
			action = "desabilitado"
		}
		m.showMessage(fmt.Sprintf("Cliente %s %s.", m.clientAction.name, action), true, clientsListScreen)
	}
	return m, nil
}

func (m Model) updateMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		next := m.returnTo
		switch next {
		case clientsListScreen:
			m.enterClients()
		case menuScreen:
			m.current = menuScreen
		default:
			m.current = next
		}
	}
	return m, nil
}

func (m *Model) enterStatus() {
	status, err := m.service.InspectStatus()
	if err != nil {
		m.showMessage(err.Error(), false, menuScreen)
		return
	}
	m.status = status
	m.current = statusScreen
}

func (m *Model) enterSetup() {
	if !m.service.IsRoot() {
		m.showMessage("Setup requer root. Rode `sudo minion ui --section setup`.", false, menuScreen)
		return
	}

	status, err := m.service.InspectStatus()
	if err != nil {
		m.showMessage(err.Error(), false, menuScreen)
		return
	}
	m.status = status
	m.setupDraft = admin.SetupOptions{
		ClientName: "default",
		ClientIPs:  "127.0.0.1/32",
	}
	m.inputs = []textinput.Model{
		newInput("Nome do cliente bootstrap", m.setupDraft.ClientName),
		newInput("IPs/CIDRs permitidos", m.setupDraft.ClientIPs),
	}
	m.inputIndex = 0
	m.focusInput(0)
	m.current = setupFormScreen
}

func (m *Model) enterConfig() {
	if !m.service.IsRoot() {
		m.showMessage("Configuracao requer root. Rode `sudo minion ui --section config`.", false, menuScreen)
		return
	}

	cfg, err := m.service.ReadConfigOrDefault()
	if err != nil {
		m.showMessage(err.Error(), false, menuScreen)
		return
	}
	m.configDraft = admin.ConfigUpdate{
		Bind:              cfg.API.Bind,
		DBPath:            cfg.DBPath,
		AllowInsecureHTTP: cfg.API.AllowInsecureHTTP,
	}
	allowInsecure := "false"
	if cfg.API.AllowInsecureHTTP {
		allowInsecure = "true"
	}
	m.inputs = []textinput.Model{
		newInput("Bind", cfg.API.Bind),
		newInput("DB path", cfg.DBPath),
		newInput("Allow insecure HTTP (true/false)", allowInsecure),
	}
	m.inputIndex = 0
	m.focusInput(0)
	m.current = configFormScreen
}

func (m *Model) enterClients() {
	if !m.service.IsRoot() {
		m.showMessage("Gerenciamento de clientes requer root. Rode `sudo minion ui --section clients`.", false, menuScreen)
		return
	}

	clients, err := m.service.ListClients()
	if err != nil {
		m.showMessage(err.Error(), false, menuScreen)
		return
	}
	m.clients = clients
	if m.clientIndex >= len(m.clients) && len(m.clients) > 0 {
		m.clientIndex = len(m.clients) - 1
	}
	if len(m.clients) == 0 {
		m.clientIndex = 0
	}
	m.current = clientsListScreen
}

func (m *Model) enterClientCreate() {
	m.clientDraft = struct {
		name string
		ips  string
	}{ips: "127.0.0.1/32"}
	m.inputs = []textinput.Model{
		newInput("Nome do cliente", ""),
		newInput("IPs/CIDRs permitidos", m.clientDraft.ips),
	}
	m.inputIndex = 0
	m.focusInput(0)
	m.current = clientCreateFormScreen
}

func (m *Model) showMessage(message string, ok bool, returnTo screen) {
	m.message = message
	m.messageOK = ok
	m.returnTo = returnTo
	m.current = messageScreen
}

func newInput(placeholder, value string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.SetValue(value)
	input.CharLimit = 256
	input.Width = 48
	return input
}

func (m *Model) focusInput(index int) {
	for i := range m.inputs {
		if i == index {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusStyle
			m.inputs[i].TextStyle = focusStyle
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m Model) viewMenu() string {
	var lines []string
	lines = append(lines, titleStyle.Render("Minion UI"))
	lines = append(lines, "")
	lines = append(lines, "Escolha uma acao:")
	for i, item := range m.menuItems {
		cursor := "  "
		if i == m.menuIndex {
			cursor = "> "
		}
		line := cursor + item
		if i == m.menuIndex {
			line = focusStyle.Render(line)
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, mutedStyle.Render("Use setas e Enter. Ctrl+C sai."))
	return strings.Join(lines, "\n")
}

func (m Model) viewStatus() string {
	lines := []string{
		titleStyle.Render("Status resumido"),
		"",
		fmt.Sprintf("Config: %s", formatBoolPath(m.status.ConfigExists, m.status.ConfigPath)),
		fmt.Sprintf("DB: %s", formatBoolPath(m.status.DBExists, m.status.DBPath)),
		fmt.Sprintf("TLS cert: %s", formatBoolPath(m.status.TLSCertExists, m.status.TLSCertPath)),
		fmt.Sprintf("TLS key: %s", formatBoolPath(m.status.TLSKeyExists, m.status.TLSKeyPath)),
		fmt.Sprintf("Clientes: %d", m.status.ClientCount),
		fmt.Sprintf("Servico: %s", m.status.ServiceStatus),
		"",
		mutedStyle.Render("r atualiza, Esc volta"),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewSetupForm() string {
	lines := []string{
		titleStyle.Render("Setup inicial"),
		"",
		fmt.Sprintf("Config atual: %s", formatBoolPath(m.status.ConfigExists, m.status.ConfigPath)),
		fmt.Sprintf("DB atual: %s", formatBoolPath(m.status.DBExists, m.status.DBPath)),
		fmt.Sprintf("TLS atual: cert=%t key=%t", m.status.TLSCertExists, m.status.TLSKeyExists),
		"",
		m.inputs[0].View(),
		m.inputs[1].View(),
		"",
		mutedStyle.Render("Tab/Shift+Tab ou setas para navegar. Enter avanca. Esc volta."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewSetupConfirm() string {
	lines := []string{
		titleStyle.Render("Confirmar setup"),
		"",
		fmt.Sprintf("Cliente bootstrap: %s", m.setupDraft.ClientName),
		fmt.Sprintf("IPs permitidos: %s", m.setupDraft.ClientIPs),
		fmt.Sprintf("Config sera garantida em: %s", m.status.ConfigPath),
		fmt.Sprintf("TLS sera garantido em: %s", filepathOrFallback(m.status.TLSCertPath, filepath.Join(m.service.TLSDir, "minion.crt"))),
		"",
		mutedStyle.Render("Enter aplica. Esc volta."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewConfigForm() string {
	lines := []string{
		titleStyle.Render("Configuracao do servico"),
		"",
		m.inputs[0].View(),
		m.inputs[1].View(),
		m.inputs[2].View(),
		"",
		mutedStyle.Render("Bind, caminho do DB e allow_insecure_http. Enter avanca, Esc volta."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewConfigConfirm() string {
	lines := []string{
		titleStyle.Render("Confirmar configuracao"),
		"",
		fmt.Sprintf("Bind: %s", m.configDraft.Bind),
		fmt.Sprintf("DB path: %s", m.configDraft.DBPath),
		fmt.Sprintf("Allow insecure HTTP: %t", m.configDraft.AllowInsecureHTTP),
		"",
		mutedStyle.Render("Enter salva. Esc volta."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewClients() string {
	lines := []string{
		titleStyle.Render("Clientes"),
		"",
	}
	if len(m.clients) == 0 {
		lines = append(lines, "Nenhum cliente cadastrado.")
	} else {
		for i, client := range m.clients {
			cursor := "  "
			if i == m.clientIndex {
				cursor = "> "
			}
			status := "desabilitado"
			if client.Enabled {
				status = "habilitado"
			}
			line := fmt.Sprintf("%s%s [%s] %s", cursor, client.Name, status, strings.Join(client.AllowedIPs, ","))
			if i == m.clientIndex {
				line = focusStyle.Render(line)
			}
			lines = append(lines, line)
		}
	}
	lines = append(lines, "")
	lines = append(lines, mutedStyle.Render("a cria, e habilita/desabilita, d remove, r atualiza, Esc volta"))
	return strings.Join(lines, "\n")
}

func (m Model) viewClientCreateForm() string {
	lines := []string{
		titleStyle.Render("Criar cliente"),
		"",
		m.inputs[0].View(),
		m.inputs[1].View(),
		"",
		mutedStyle.Render("Enter avanca. Esc cancela."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewClientCreateConfirm() string {
	lines := []string{
		titleStyle.Render("Confirmar criacao de cliente"),
		"",
		fmt.Sprintf("Nome: %s", m.clientDraft.name),
		fmt.Sprintf("IPs permitidos: %s", m.clientDraft.ips),
		"",
		mutedStyle.Render("Enter cria. Esc volta."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewClientActionConfirm() string {
	action := "desabilitar"
	if m.clientAction.delete {
		action = "remover"
	} else if m.clientAction.enabled {
		action = "habilitar"
	}
	lines := []string{
		titleStyle.Render("Confirmar acao em cliente"),
		"",
		fmt.Sprintf("Acao: %s", action),
		fmt.Sprintf("Cliente: %s", m.clientAction.name),
		"",
		mutedStyle.Render("Enter confirma. Esc volta."),
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewMessage() string {
	style := errorStyle
	if m.messageOK {
		style = okStyle
	}
	return strings.Join([]string{
		style.Render("Resultado"),
		"",
		m.message,
		"",
		mutedStyle.Render("Enter volta."),
	}, "\n")
}

func formatBoolPath(exists bool, path string) string {
	if exists {
		return path
	}
	return path + " (ausente)"
}

func filepathOrFallback(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
