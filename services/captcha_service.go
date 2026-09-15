package services

import (
	"errors"
	"sync"
	"time"

	"hive-admin-go/utils"
)

const (
	// captchaChallengeTTL 挑战票据有效期，过期未消费即作废
	captchaChallengeTTL = 5 * time.Minute
	// captchaIssueWindow 与 captchaIssueLimitPerIP 签发限流：同一 IP 在窗口内最多签发次数
	captchaIssueWindow     = time.Minute
	captchaIssueLimitPerIP = 10
)

var (
	// ErrCaptchaRequired 登录触发防爆破校验但未携带有效挑战（缺失、已消费或已过期）
	ErrCaptchaRequired = errors.New("请完成滑块验证后重试")
	// ErrCaptchaTooFrequent 挑战签发超出同 IP 频率上限
	ErrCaptchaTooFrequent = errors.New("操作过于频繁，请稍后再试")
)

// captchaStore 进程内共享的挑战票据与签发计数。单实例内存实现：
// 重启丢失仅要求用户重新滑动一次，无持久化必要（见 ADR 0005）。
type captchaStore struct {
	mu         sync.Mutex
	challenges map[string]time.Time // captchaId -> 签发时刻，消费即删除
	issued     map[string][]time.Time
}

var sharedCaptchaStore = &captchaStore{
	challenges: make(map[string]time.Time),
	issued:     make(map[string][]time.Time),
}

// CaptchaService 负责滑块挑战票据的签发、限流与一次性消费。
type CaptchaService struct {
	store *captchaStore
}

func NewCaptchaService() *CaptchaService {
	return &CaptchaService{store: sharedCaptchaStore}
}

// Issue 为来源 IP 签发一枚一次性挑战票据，超出同 IP 窗口签发上限时返回 ErrCaptchaTooFrequent。
func (s *CaptchaService) Issue(ip string) (string, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	now := time.Now()
	recent := s.store.issued[ip][:0]
	for _, issuedAt := range s.store.issued[ip] {
		if now.Sub(issuedAt) < captchaIssueWindow {
			recent = append(recent, issuedAt)
		}
	}
	if len(recent) >= captchaIssueLimitPerIP {
		s.store.issued[ip] = recent
		return "", ErrCaptchaTooFrequent
	}
	s.store.issued[ip] = append(recent, now)

	// 惰性清理：挑战数量超限时删除过期票据，静默 IP 的签发计数随窗口过滤与批量清理回收
	if len(s.store.challenges) > 1024 {
		for id, issuedAt := range s.store.challenges {
			if now.Sub(issuedAt) > captchaChallengeTTL {
				delete(s.store.challenges, id)
			}
		}
	}
	if len(s.store.issued) > 4096 {
		for key, timestamps := range s.store.issued {
			alive := false
			for _, issuedAt := range timestamps {
				if now.Sub(issuedAt) < captchaIssueWindow {
					alive = true
					break
				}
			}
			if !alive {
				delete(s.store.issued, key)
			}
		}
	}

	id := utils.GenerateUUID()
	s.store.challenges[id] = now
	return id, nil
}

// Consume 原子消费一枚挑战票据：票据缺失、已消费或已过期时返回 ErrCaptchaRequired。
// 无论结果如何票据都被消费，不可重放。不校验拖动耗时：脚本可以谎报任意耗时，
// 该校验无真实防护价值，反而误伤快速拖动的真人用户。
func (s *CaptchaService) Consume(id string) error {
	if id == "" {
		return ErrCaptchaRequired
	}

	s.store.mu.Lock()
	issuedAt, ok := s.store.challenges[id]
	if ok {
		delete(s.store.challenges, id)
	}
	s.store.mu.Unlock()

	if !ok || time.Since(issuedAt) > captchaChallengeTTL {
		return ErrCaptchaRequired
	}
	return nil
}
