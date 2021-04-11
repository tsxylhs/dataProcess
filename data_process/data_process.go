package data_process

import "fmt"

//   1，初始化任务的goroutine
//
//   2，分发任务的goroutine
//
//   3，等到所有work结束，然后关闭所有通道的goroutine
//
//main主要负责拉起以上的goroutine 并取结果
//
// 程序还需要三个通道
//
//  1，传递task任务的通道
//
//  2，传递task结果的通道
//
// 3 ，接收workder处理完任务后所发送通知的通道
//
//具体代码如下
//定义工作数量
const (
	WORKS = 5
)

//定义工作任务结构体，可根据需求改变
type task struct {
	begin  int
	end    int
	result chan<- int
}

//定义执行任务的方法，可根据需求更改
func (t *task) do() {
	sum := 0
	for i := t.begin; i <= t.end; i++ {
		sum += i
	}
	t.result <- sum
}

//入口函数

func DataProcess() {
	works := WORKS
	//定义工作通道
	taskchan := make(chan task, 10)
	//定义结果通道
	resultchan := make(chan int, 10)
	//work工作信号通道
	done := make(chan struct{}, 10)
	//初始化task的goroutine
	go initTask(taskchan, resultchan, 100)
	//分发任务到协程池
	distributeTask(taskchan, works, done)
	//获取goroutine处理完成任务通知，并关闭通道
	go closeResult(done, resultchan, works)
	//通过结果通道，获取结果并汇总
	sum := processResult(resultchan)
	fmt.Println("sum=", sum)
}

//初始化task chan
func initTask(taskchan chan<- task, r chan int, p int) {
	qu := p / 10
	mod := p % 10
	high := qu * 10
	for j := 0; j < qu; j++ {
		b := 10*j + 1
		e := 10 * (j + 1)
		task := task{
			begin:  b,
			end:    e,
			result: r,
		}
		taskchan <- task
	}
	if mod != 0 {
		task := task{
			begin:  high + 1,
			end:    p,
			result: r,
		}
		taskchan <- task
	}
	close(taskchan)
}

//读取taskchan 并分发到worker goRoutine处理，总数量为workers
func distributeTask(taskchan <-chan task, workers int, done chan struct{}) {
	for i := 0; i < workers; i++ {
		go processTask(taskchan, done)
	}
}

//工作goroutine处理的具体内容，并将处理的结果发送到结果chan
func processTask(taskchan <-chan task, done chan struct{}) {
	for t := range taskchan {
		t.do()
	}
	done <- struct{}{}
}

//通过done channel同步等待所有工作goroutine的结束，然后关闭结果chan
func closeResult(done chan struct{}, resultchan chan int, workers int) {
	for i := 0; i < workers; i++ {
		<-done
	}
	close(done)
	close(resultchan)
}

//读取结果通道汇聚结果

func processResult(resultchan chan int) int {
	sum := 0
	for r := range resultchan {
		sum += r
	}
	return sum
}
