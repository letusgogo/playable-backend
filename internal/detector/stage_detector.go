package detector

import (
	"context"
)

type StageChecker interface {
	// 传入截图（整图或多区域），返回判定阶段以及命中细节
	Detect(ctx context.Context, game string, currentStageNum int, imgBase64 string) (match bool, evidence string, err error)
}
