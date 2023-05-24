package main

import (
	"demo2/pkg/utils/loggers"
	"fmt"
	"time"
)

type State int

/*
状态：0 = Pending
状态：1 = Deploying
状态：2 = StartupSuccess
状态：3 = Failure
状态：4 = Revoked
*/

const (
	Pending          State = iota //等待
	Deploying                     //部署中
	StartupSuccess                //启动探针探测成功
	ReadinessSuccess              //就绪探针探测成功
	Failure                       //失败
	Revoked                       //撤销
)

type Event int

const (
	Event1              Event = iota // 定义事件1
	Event2                           // 定义事件2
	Event3                           // 定义事件3
	EventStartupDeploy               //事件：开始进行部署
	EventStartupProbe                //事件：开始进行启动探测
	EventReadinessProbe              //事件：开始启动就绪探测
)

type EventAction interface {
	Exec() bool
}

type StartupDeploy struct{}

func (s *StartupDeploy) Exec() bool {
	loggers.DefaultLogger.Info("开始启动部署...")
	return true
}

type StartupProbe struct{}

func (s *StartupProbe) Exec() bool {
	loggers.DefaultLogger.Info("启动探针开始探测...")
	return true
}

type ReadinessProbe struct{}

func (s *ReadinessProbe) Exec() bool {
	loggers.DefaultLogger.Info("就绪探针开始探测...")
	return false
} // Transition 过渡链
type Transition struct {
	CurrentState State // 当前状态
	Event        Event // 触发事件
	NextState    State // 下一个状态
}

// StateMachine 状态机
type StateMachine struct {
	Transitions  []Transition // 状态转移数组
	CurrentState State        // 当前状态
}

// Trigger 触发器
func (sm *StateMachine) Trigger(event Event) { // 触发状态转移
	for _, t := range sm.Transitions {
		if t.CurrentState == sm.CurrentState && t.Event == event {
			sm.CurrentState = t.NextState
			return
		}
	}
}

// RollingEngine 滚动引擎
func RollingEngine(schedulingChain []Transition, CurrentState State, ExecEvent Event) State {
	//初始化状态机
	fsm := StateMachine{
		Transitions:  schedulingChain,
		CurrentState: CurrentState,
	}

	//根据【触发事件】得到【下一个状态】
	fsm.Trigger(ExecEvent)

	//返回新状态给调用方
	return fsm.CurrentState
}

func main() {
	//创建调度链
	schedulingChain := []Transition{
		{Pending, Event1, Deploying},        // 状态转移链1
		{Deploying, Event2, StartupSuccess}, // 状态转移链2
		{StartupSuccess, Event3, Pending},   // 状态转移链3
		{Pending, EventStartupDeploy, Deploying},
		{Deploying, EventStartupProbe, StartupSuccess},
		{StartupSuccess, EventReadinessProbe, ReadinessSuccess},
		{Pending, Event2, Failure},
		{Deploying, Event3, Revoked},
	}

	//创建一个焦点状态（默认为PENDING）
	FocusState := Pending
	var ea EventAction
	var nums = 0

	for {
		switch FocusState {
		case Pending:
			ea = &StartupDeploy{}
			if ea.Exec() {
				FocusState = RollingEngine(schedulingChain, Pending, EventStartupDeploy)
			}

		case Deploying:
			ea = &StartupProbe{}
			if ea.Exec() {
				FocusState = RollingEngine(schedulingChain, Deploying, EventStartupProbe)
			}

		case StartupSuccess:
			ea = &ReadinessProbe{}
			if ea.Exec() {
				FocusState = RollingEngine(schedulingChain, StartupSuccess, EventReadinessProbe)
			}
			time.Sleep(2 * time.Second)
			nums++

			if nums > 10 {
				FocusState = RollingEngine(schedulingChain, StartupSuccess, EventReadinessProbe)
			}

		case ReadinessSuccess:
			loggers.DefaultLogger.Info("就绪探测成功，服务已经部署成功....")
			goto breakHere
		}
	}

breakHere:
	fmt.Println("跳出重试循环...")
}

