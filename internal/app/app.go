package app

// App 是应用内核的语义别名，保留调用侧 API 语义。
type App = Kernel

// NewApp 提供应用层构造入口，保持 main/newApp 语义清晰。
func NewApp(webComponents []Component, sshComponent Component) *App {
	return NewKernel(webComponents, sshComponent)
}
