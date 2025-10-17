package utils

import "strings"

// IsRealError 智能判断stderr输出是否为真正的错误
func IsRealError(line string) bool {
	line = strings.TrimSpace(strings.ToLower(line))

	// Git命令的正常信息，不应该被标记为错误
	gitNormalPatterns := []string{
		"already on",                // Already on 'branch'
		"up to date",                // Already up to date
		"your branch is up to date", // Your branch is up to date
		"your branch is ahead",      // Your branch is ahead
		"your branch is behind",     // Your branch is behind
		"switching to",              // Switching to branch
		"branch",                    // Branch related messages
		"checkout",                  // Checkout messages
		"reset",                     // Reset messages
		"clean",                     // Clean messages
		"stash",                     // Stash messages
		"pull",                      // Pull messages
		"push",                      // Push messages
		"commit",                    // Commit messages
		"merge",                     // Merge messages
		"fast-forward",              // Fast-forward messages
		"downloading",               // Downloading messages
		"receiving objects",         // Receiving objects
		"resolving deltas",          // Resolving deltas
		"compressing objects",       // Compressing objects
		"writing objects",           // Writing objects
		"counting objects",          // Counting objects
		"remote:",                   // Remote messages
		"origin/",                   // Origin branch references
		"refs/heads/",               // Branch references
		"refs/remotes/",             // Remote references
	}

	// 检查是否匹配Git正常信息模式
	for _, pattern := range gitNormalPatterns {
		if strings.Contains(line, pattern) {
			return false // 不是错误，是正常信息
		}
	}

	// Yarn/NPM的正常信息
	yarnNormalPatterns := []string{
		"yarn install",
		"yarn build",
		"info",
		"success",
		"done in",
		"warning",    // 警告不是错误
		"deprecated", // 废弃警告不是错误
		"peer dep",   // peer dependency警告
		"optional",   // 可选依赖警告
	}

	for _, pattern := range yarnNormalPatterns {
		if strings.Contains(line, pattern) {
			return false
		}
	}

	// 构建工具的正常信息
	buildNormalPatterns := []string{
		"webpack",
		"compiled successfully",
		"build complete",
		"assets by path",
		"chunk",
		"modules by path",
		"orphan modules",
		"runtime modules",
		"cached modules",
		"built at:",
		"hash:",
		"version:",
		"time:",
		"size:",
	}

	for _, pattern := range buildNormalPatterns {
		if strings.Contains(line, pattern) {
			return false
		}
	}

	// 真正的错误模式
	errorPatterns := []string{
		"error:",
		"fatal:",
		"failed:",
		"cannot",
		"unable to",
		"permission denied",
		"no such file",
		"command not found",
		"syntax error",
		"compilation failed",
		"build failed",
		"test failed",
		"timeout",
		"killed",
		"segmentation fault",
		"core dumped",
		"exception",
		"traceback",
		"panic:",
	}

	for _, pattern := range errorPatterns {
		if strings.Contains(line, pattern) {
			return true // 这是真正的错误
		}
	}

	// 如果都不匹配，默认认为不是错误（更保守的策略）
	return false
}
