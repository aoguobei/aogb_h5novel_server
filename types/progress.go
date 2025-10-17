package types

// ProgressCallback 进度回调函数类型
type ProgressCallback func(percentage int, text string, detail string)

// ProgressContext 进度上下文
type ProgressContext struct {
	callback ProgressCallback
	startPct int
	rangePct int
}

// NewProgressContext 创建进度上下文
func NewProgressContext(callback ProgressCallback, startPct, rangePct int) *ProgressContext {
	return &ProgressContext{
		callback: callback,
		startPct: startPct,
		rangePct: rangePct,
	}
}

// UpdateStepProgress 更新步骤进度
func (pc *ProgressContext) UpdateStepProgress(stepPct int, text string, detail string) {
	if pc.callback != nil {
		percentage := pc.startPct + (stepPct * pc.rangePct / 100)
		pc.callback(percentage, text, detail)
	}
}
