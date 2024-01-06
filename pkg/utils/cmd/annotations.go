package cmd

const (
	// AnnotationRunAsRoot 标记需要以 root 用户运行的注解
	AnnotationRunAsRoot = "run-as-root"
	// AnnotationRequireManager 标记需要 manager.Manager 的注解
	AnnotationRequireManager = "require-manager"

	// AnnotationValueTrue 表示逻辑“真”的注解值
	AnnotationValueTrue = "true"
)
