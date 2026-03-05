package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	chssh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/tui"
)

// Server 是 SSH runtime 的实例化封装。
type Server struct {
	server *chssh.Server
}

// New 创建一个 SSH runtime。
func New() *Server {
	return &Server{}
}

// Start 启动 SSH 服务器。
func (s *Server) Start(context.Context) error {
	if s.server != nil {
		return errors.New("ssh server already running")
	}

	host := config.Config().SSH.Host
	port := config.Config().SSH.Port
	key := config.Config().SSH.Key

	newServer, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(key),
		wish.WithMiddleware(
			BubbleteaMiddleware(teaHandler),
			ActivetermMiddleware(), // Bubble Tea apps usually require a PTY.
		),
	)
	if err != nil {
		return err
	}

	s.server = newServer

	go func() {
		fmt.Println("🚀 Ech0 SSH已启动，监听端口", port)
		if serveErr := s.server.ListenAndServe(); serveErr != nil &&
			!errors.Is(serveErr, chssh.ErrServerClosed) {
			fmt.Printf("ssh server run failed: %v\n", serveErr)
		}
	}()

	return nil
}

// Stop 停止 SSH 服务器。
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil && !errors.Is(err, chssh.ErrServerClosed) {
		// 强制关闭服务器
		if closeErr := s.server.Close(); closeErr != nil {
			fmt.Printf("force close ssh server failed: %v\n", closeErr)
		}

		return err
	}

	s.server = nil
	return nil
}

// IsRunning 返回 SSH runtime 是否运行中。
func (s *Server) IsRunning() bool {
	return s.server != nil
}

// ActivetermMiddleware Middleware will exit 1 connections trying with no active terminals.
func ActivetermMiddleware() wish.Middleware {
	return func(next chssh.Handler) chssh.Handler {
		return func(sess chssh.Session) {
			_, _, active := sess.Pty()
			if active {
				next(sess)
				return
			}
			wish.Println(sess, "Requires an active PTY")
			_ = sess.Exit(1)
		}
	}
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// tea.WithAltScreen) on a session by session basis.
func teaHandler(s chssh.Session) (tea.Model, []tea.ProgramOption) {
	// This should never fail, as we are using the activeterm middleware.
	pty, _, _ := s.Pty()

	// When running a Bubble Tea app over SSH, you shouldn't use the default
	// lipgloss.NewStyle function.
	// That function will use the color profile from the os.Stdin, which is the
	// server, not the client.
	// We provide a MakeRenderer function in the bubbletea middleware package,
	// so you can easily get the correct renderer for the current session, and
	// use it to create the styles.
	// The recommended way to use these styles is to then pass them down to
	// your Bubble Tea model.
	renderer := MakeRenderer(s)
	txtStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	quitStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))

	bg := "light"
	if renderer.HasDarkBackground() {
		bg = "dark"
	}

	m := model{
		term:      pty.Term,
		profile:   renderer.ColorProfile().Name(),
		width:     pty.Window.Width,
		height:    pty.Window.Height,
		bg:        bg,
		txtStyle:  txtStyle,
		quitStyle: quitStyle,
		logo:      tui.GetLogoBanner(),
		textarea:  textarea.New(),
	}

	m.textarea.Placeholder = "请输入..."
	m.textarea.Focus()

	return m, []tea.ProgramOption{
		tea.WithAltScreen(),
	}
}

// model TUI 模型定义
type model struct {
	term      string
	profile   string
	width     int
	height    int
	bg        string
	txtStyle  lipgloss.Style
	quitStyle lipgloss.Style
	textarea  textarea.Model
	logo      string
}

// Init 初始化TUI
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update 更新View的内容
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // 声明 cmds 切片
	var cmd tea.Cmd    // 声明 cmd 变量

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 处理窗口大小变化消息。当 SSH 客户端窗口大小改变时，wish 中间件会发送此消息。
		m.height = msg.Height
		m.width = msg.Width
		// 将消息传递给 textarea，让它也能调整自身大小（如果需要）
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd) // 将 textarea 返回的命令添加到列表中
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// 处理 'Ctrl+C' 退出命令
			return m, tea.Quit
		case "esc": // 添加对 Esc 键的处理
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		default:
			// 如果 textarea 没有焦点，按下任意键使其获得焦点
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
		// 将输入的消息强制转换编码为 UTF-8
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd) // 将 textarea 返回的命令添加到列表中
	}

	// 返回更新后的模型和所有累积的命令
	return m, tea.Batch(cmds...)
}

// View 渲染TUI页面内容
func (m model) View() string {
	return tui.GetSSHView()
}
