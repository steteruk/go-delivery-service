package workerpool

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/steteruk/go-delivery-service/location/domain"
)

// LocationPool add count tasks in courierLocationQueue for handling these tasks and run count workers countWorkers
// It needs when we have a lot of requests.
type LocationPool struct {
	courierLocationQueue    chan *domain.CourierLocation
	courierService          domain.CourierLocationServiceInterface
	countTasks              int
	countWorkers            int
	isClosed                bool
	mu                      sync.Mutex
	timeoutGracefulShutdown time.Duration
}

// Run inits workerPools define count task and count workers.
func (wl *LocationPool) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	wl.courierLocationQueue = make(chan *domain.CourierLocation, wl.countTasks)
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < wl.countWorkers; i++ {
		go wl.handleTasks(cancelCtx)
	}

	<-ctx.Done()

	wl.isClosed = true
	close(wl.courierLocationQueue)
	wl.gracefulShutdown()
}

func (wl *LocationPool) handleTasks(ctx context.Context) {
	for courierLocation := range wl.courierLocationQueue {
		select {
		case <-ctx.Done():
			return
		default:
			err := wl.courierService.SaveLatestCourierLocation(ctx, courierLocation)
			if err != nil {
				log.Printf("failed to save latest position: %v\n", err)
			}
		}
	}
}

func (wl *LocationPool) gracefulShutdown() {
	timer := time.After(wl.timeoutGracefulShutdown)
	for {
		select {
		case <-timer:
			return
		default:

			if len(wl.courierLocationQueue) == 0 {
				return
			}

		}
	}
}

// AddTask adds task in LocationQueue.
func (wl *LocationPool) AddTask(courierLocation *domain.CourierLocation) {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	if !wl.isClosed {
		wl.courierLocationQueue <- courierLocation
	}
}

// NewLocationPool creates new worker pools.
func NewLocationPool(
	courierLocationService domain.CourierLocationServiceInterface,
	countWorkers int,
	countTasks int,
	timeoutGracefulShutdown time.Duration,
) *LocationPool {
	return &LocationPool{
		courierService:          courierLocationService,
		countWorkers:            countWorkers,
		countTasks:              countTasks,
		timeoutGracefulShutdown: timeoutGracefulShutdown,
	}
}
