// worker
package main

import (
	"container/heap"
	"priceloader"
	"sync"
)

//Рабочий
type Worker struct {
	task    chan priceloader.LoadTask // канал для заданий
	pending int                       // кол-во оставшихся задач
	index   int                       // позиция в куче
	wg      *sync.WaitGroup           //указатель на группу ожидания
}

func (w *Worker) work(done chan *Worker) {
	for {
		taskToDo := <-w.task  //читаем следующее задание
		w.wg.Add(1)           //инкриминируем счетчик группы ожидания
		getItemHtml(taskToDo) //загружаем позиции
		w.wg.Done()           //сигнализируем группе ожидания что закончили
		done <- w             //показываем что завершили работу
	}
}

//Это будет наша "куча":
type Pool []*Worker

//Проверка кто меньше - в нашем случае меньше тот у кого меньше заданий:
func (p Pool) Less(i, j int) bool { return p[i].pending < p[j].pending }

//Вернем количество рабочих в пуле:
func (p Pool) Len() int { return len(p) }

//Реализуем обмен местами:
func (p Pool) Swap(i, j int) {
	if i >= 0 && i < len(p) && j >= 0 && j < len(p) {
		p[i], p[j] = p[j], p[i]
		p[i].index, p[j].index = i, j
	}
}

//Заталкивание элемента:
func (p *Pool) Push(x interface{}) {
	n := len(*p)
	worker := x.(*Worker)
	worker.index = n
	*p = append(*p, worker)
}

//И выталкивание:
func (p *Pool) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	item.index = -1
	*p = old[0 : n-1]
	return item
}

//Балансировщик
type Balancer struct {
	pool     Pool                      //Наша "куча" рабочих
	done     chan *Worker              //Канал уведомления о завершении для рабочих
	requests chan priceloader.LoadTask //Канал для получения новых заданий
	flowctrl chan bool                 //Канал для PMFC
	queue    int                       //Количество незавершенных заданий переданных рабочим
	wg       *sync.WaitGroup           //Группа ожидания для рабочих
}

//Инициализируем балансировщик. Аргументом получаем канал по которому приходят задания
func (b *Balancer) init(task chan priceloader.LoadTask) {
	b.requests = make(chan priceloader.LoadTask)
	b.flowctrl = make(chan bool)
	b.done = make(chan *Worker)
	b.wg = new(sync.WaitGroup)

	//Запускаем наш Flow Control:
	go func() {
		for {
			b.requests <- <-task //получаем новое задание и пересылаем его на внутренний канал
			<-b.flowctrl         //а потом ждем получения подтверждения
		}
	}()

	//Инициализируем кучу и создаем рабочих:
	heap.Init(&b.pool)
	for i := 0; i < WORKERS; i++ {
		w := &Worker{
			task:    make(chan priceloader.LoadTask, WORKERSCAP),
			index:   0,
			pending: 0,
			wg:      b.wg,
		}
		go w.work(b.done)     //запускаем рабочего
		heap.Push(&b.pool, w) //и заталкиваем его в кучу
	}
}

//Рабочая функция балансировщика получает аргументом канал уведомлений от главного цикла
func (b *Balancer) balance(quit chan bool) {
	lastjobs := false //Флаг завершения, поднимаем когда кончились задания
	for {
		select { //В цикле ожидаем коммуникации по каналам:

		case <-quit: //пришло указание на остановку работы
			b.wg.Wait()  //ждем завершения текущих загрузок рабочими..
			quit <- true //..и отправляем сигнал что закончили

		case task := <-b.requests: //Получено новое задание (от flow controller)
			if task.Message != ENDMESSAGE { //Проверяем - а не кодовая ли это фраза?
				b.dispatch(task) // если нет, то отправляем рабочим
			} else {
				lastjobs = true //иначе поднимаем флаг завершения
			}

		case w := <-b.done: //пришло уведомление, что рабочий закончил загрузку
			b.completed(w) //обновляем его данные
			if lastjobs {
				if w.pending == 0 { //если у рабочего кончились задания..
					heap.Remove(&b.pool, w.index) //то удаляем его из кучи
				}
				if len(b.pool) == 0 { //а если куча стала пуста
					//значит все рабочие закончили свои очереди
					quit <- true //и можно отправлять сигнал подтверждения готовности к останову
				}
			}
		}
	}
}

// Функция отправки задания
func (b *Balancer) dispatch(task priceloader.LoadTask) {
	w := heap.Pop(&b.pool).(*Worker) //Берем из кучи самого незагруженного рабочего..
	w.task <- task                   //..и отправляем ему задание.
	w.pending++                      //Добавляем ему "весу"..
	heap.Push(&b.pool, w)            //..и отправляем назад в кучу
	if b.queue++; b.queue < WORKERS*WORKERSCAP {
		b.flowctrl <- true
	}
}

//Обработка завершения задания
func (b *Balancer) completed(w *Worker) {
	w.pending--
	heap.Remove(&b.pool, w.index)
	heap.Push(&b.pool, w)
	if b.queue--; b.queue == WORKERS*WORKERSCAP-1 {
		b.flowctrl <- true
	}
}
