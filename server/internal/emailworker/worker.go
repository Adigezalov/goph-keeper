package emailworker

import (
	"context"
	"sync"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/logger"
)

type EmailJob struct {
	To      string
	Code    string
	Attempt int
}

type EmailSender interface {
	SendVerificationCode(toEmail, code string) error
}

type Worker struct {
	emailService EmailSender
	jobQueue     chan EmailJob
	maxRetries   int
	retryDelay   time.Duration
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewWorker(emailService EmailSender, queueSize int, maxRetries int, retryDelay time.Duration) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		emailService: emailService,
		jobQueue:     make(chan EmailJob, queueSize),
		maxRetries:   maxRetries,
		retryDelay:   retryDelay,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go w.run()
	logger.Info("[EmailWorker] Worker запущен")
}

func (w *Worker) Stop() {
	logger.Info("[EmailWorker] Остановка worker...")
	w.cancel()
	close(w.jobQueue)
	w.wg.Wait()
	logger.Info("[EmailWorker] Worker остановлен")
}

func (w *Worker) SendEmail(to, code string) {
	w.sendEmailWithAttempt(to, code, 0)
}

func (w *Worker) sendEmailWithAttempt(to, code string, attempt int) {
	job := EmailJob{
		To:      to,
		Code:    code,
		Attempt: attempt,
	}

	select {
	case w.jobQueue <- job:
		if attempt == 0 {
			logger.Infof("[EmailWorker] Задача добавлена в очередь: %s", to)
		}
	case <-w.ctx.Done():
		logger.Warn("[EmailWorker] Worker остановлен, задача отклонена")
	default:
		logger.Warn("[EmailWorker] Очередь заполнена, задача отклонена")
	}
}

func (w *Worker) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("[EmailWorker] Получен сигнал остановки")
			// Обрабатываем оставшиеся задачи
			w.drainQueue()
			return
		case job, ok := <-w.jobQueue:
			if !ok {
				return
			}
			w.processJob(job)
		}
	}
}

func (w *Worker) processJob(job EmailJob) {
	err := w.emailService.SendVerificationCode(job.To, job.Code)
	if err != nil {
		logger.Errorf("[EmailWorker] Ошибка отправки email на %s (попытка %d): %v", job.To, job.Attempt+1, err)

		if job.Attempt < w.maxRetries {
			logger.Infof("[EmailWorker] Повторная попытка через %v", w.retryDelay)
			time.Sleep(w.retryDelay)
			w.sendEmailWithAttempt(job.To, job.Code, job.Attempt+1)
		} else {
			logger.Errorf("[EmailWorker] Превышено максимальное количество попыток для %s", job.To)
		}
		return
	}

	logger.Infof("[EmailWorker] Email успешно отправлен на %s", job.To)
}

func (w *Worker) drainQueue() {
	logger.Info("[EmailWorker] Обработка оставшихся задач в очереди...")
	for {
		select {
		case job, ok := <-w.jobQueue:
			if !ok {
				return
			}
			w.processJob(job)
		case <-time.After(100 * time.Millisecond):
			return
		}
	}
}

func (w *Worker) QueueLength() int {
	return len(w.jobQueue)
}
