package main

import "C"
import (
	"flag"
	"github.com/chistiykot/gobot/window"
	"github.com/go-vgo/robotgo"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	applicationWindowName   = "New World"
	applicationProcessName  = "NewWorld.exe"
	applicationWindowWidth  = 1920
	applicationWindowHeight = 1080
)

type Direction int

const (
	forward Direction = iota
	left
	right
	any
)

const (
	runMinTime    = 1500
	runMaxTime    = 4500
	runMaxTurnGap = 30
	turnMaxGap    = 300
)

const (
	alwaysRunTickerMinTime                = 4000
	alwaysRunTickerMaxTime                = 8000
	alwaysRunActionTickerMinTime          = 800
	alwaysRunActionTickerMaxTime          = 1500
	alwaysRunDirectionChangeTickerMinTime = 1500
	alwaysRunDirectionChangeTickerMaxTime = 6000
)

const (
	//actionBitmapSource         = "action_1920x1080.png" // "E" letter image (used for action recognition)
	actionBitmapSource         = "action_side_1920x1080.png" // part of action square (used for action recognition)
	actionBitmapTolerance      = 0.4
	actionScanScreenCaptureLag = 400
	actionMinTime              = 4000
	actionMaxTime              = 9000
)

type GameWindow struct {
	pid                      int32
	hwnd                     window.HWND
	left, top, width, height int
}

type BotConfiguration struct {
	GameWindow
	actionBitmapFile      string
	actionBitmapTolerance float64
}

type Bot struct {
	config BotConfiguration
}

func (bot Bot) recognizeActionCoordinates() (int, int) {
	log.Printf("action scan start")

	actionBitmap := robotgo.OpenBitmap(bot.config.actionBitmapFile)
	screenBitmap := robotgo.CaptureScreen(bot.config.left, bot.config.top, bot.config.width, bot.config.height)
	defer robotgo.FreeBitmap(actionBitmap)
	defer robotgo.FreeBitmap(screenBitmap)

	bot.sleep(actionScanScreenCaptureLag)

	fx, fy := robotgo.FindBitmap(actionBitmap, screenBitmap, bot.config.actionBitmapTolerance)
	if fx == -1 {
		log.Printf("action scan fail: not found")
	} else {
		log.Printf("action scan succes: x, y: %d, %d", fx, fy)
	}

	return fx, fy
}

func getConfiguration() BotConfiguration {
	applicationInfo := window.GetApplicationInfo(applicationWindowName, applicationProcessName)

	left := int(applicationInfo.Left)
	top := int(applicationInfo.Top)

	return BotConfiguration{
		GameWindow{applicationInfo.Pid, applicationInfo.Hwnd, left, top, applicationWindowWidth, applicationWindowHeight},
		actionBitmapSource,
		actionBitmapTolerance,
	}
}

func (bot Bot) getRandomDirection() Direction {
	return Direction(rand.Intn(int(any)))
}

func (bot Bot) getRandomTurnDirection() Direction {
	return Direction(1 + rand.Intn(int(any-1)))
}

func (bot Bot) randomTurn(direction Direction) {
	if direction == any {
		direction = bot.getRandomTurnDirection()
	}

	turnGap := rand.Intn(turnMaxGap)
	bot.turnAround(direction, turnGap)
}

func (bot Bot) randomRun(direction Direction, time int) {
	if direction == any {
		direction = bot.getRandomDirection()
	}

	turnGap := rand.Intn(runMaxTurnGap)
	bot.run(direction, turnGap, time)
}

func (bot Bot) sleep(ms int) {
	log.Printf("sleep for %d ms", ms)

	robotgo.MilliSleep(ms)
}

func (bot Bot) getRandomRange(min, max int) int {
	return min + rand.Intn(max-min)
}

func (bot Bot) randomSleep(min, max int) {
	bot.sleep(bot.getRandomRange(min, max))
}

func (bot Bot) run(direction Direction, turnGap int, time int) {
	log.Printf("run start, direction: `%d`, run time: `%d` ms", direction, time)

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				robotgo.KeyToggle("w", "up")
				return
			default:
				if direction == left {
					bot.turnLeft(turnGap)
				} else if direction == right {
					bot.turnRight(turnGap)
				}
				robotgo.KeyToggle("w", "down")
				robotgo.MilliSleep(10)
			}
		}
	}()

	robotgo.MilliSleep(time)
	done <- true

	log.Printf("run finish")
}

func (bot Bot) alwaysRun() {
	log.Printf("always run")

	robotgo.KeyTap("=")
}

func (bot Bot) turnAround(direction Direction, gap int) {
	log.Printf("turn start, direction: `%d`, gap: `%d`", direction, gap)

	if direction == left {
		bot.turnLeft(gap)
	} else if direction == right {
		bot.turnRight(gap)
	}

	log.Printf("turn finish")
}

func (bot Bot) turnLeft(gap int) {
	cx, cy := robotgo.GetMousePos()

	robotgo.MoveMouseSmooth(cx-gap, cy)
}

func (bot Bot) turnRight(gap int) {
	cx, cy := robotgo.GetMousePos()

	robotgo.MoveMouseSmooth(cx+gap, cy)
}

func (bot Bot) invokeAction() {
	log.Printf("action invoke")

	robotgo.KeyTap("e")
}

func runTestCommandIfFlagSet(bot Bot, flag string) {
	if flag == "action" {
		if bot.recognizeAndInvokeAction() {
			bot.randomSleep(actionMinTime, actionMaxTime)
		}
	}

	if flag == "turn" {
		bot.randomTurn(any)
	}

	if flag != "" {
		os.Exit(0)
	}
}

func (bot Bot) recognizeAndInvokeAction() bool {
	actionX, _ := bot.recognizeActionCoordinates()

	if actionX == -1 {
		return false
	}

	bot.invokeAction()

	return true
}

func (bot Bot) createRandomTicker(min, max int) *time.Ticker {
	runTime := time.Duration(int32(bot.getRandomRange(min, max)))

	return time.NewTicker(runTime * time.Millisecond)
}

func (bot Bot) alwaysRunAndInvokeAction() {
	runTicker := bot.createRandomTicker(alwaysRunTickerMinTime, alwaysRunTickerMaxTime)
	actionTicker := bot.createRandomTicker(alwaysRunActionTickerMinTime, alwaysRunActionTickerMaxTime)
	randomDirectionTicker := bot.createRandomTicker(alwaysRunDirectionChangeTickerMinTime, alwaysRunDirectionChangeTickerMaxTime)

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-runTicker.C:
				bot.alwaysRun()
				runTicker.Stop()
				runTicker = bot.createRandomTicker(alwaysRunTickerMinTime, alwaysRunTickerMaxTime)
			case <-actionTicker.C:
				bot.invokeAction()
				actionTicker.Stop()
				actionTicker = bot.createRandomTicker(alwaysRunActionTickerMinTime, alwaysRunActionTickerMaxTime)
			case <-randomDirectionTicker.C:
				bot.centerView()
				bot.randomTurn(any)
				randomDirectionTicker.Stop()
				randomDirectionTicker = bot.createRandomTicker(alwaysRunDirectionChangeTickerMinTime, alwaysRunDirectionChangeTickerMaxTime)
			case <-quit:
				runTicker.Stop()
				return
			}
		}
	}()
}

func (bot Bot) randomRunAndInvokeAction() {
	runTime := rand.Intn(runMaxTime-runMinTime) + runMinTime
	bot.centerView()
	bot.randomRun(any, runTime)
	bot.sleep(800)

	for bot.recognizeAndInvokeAction() {
		bot.randomSleep(actionMinTime, actionMaxTime)
	}
}

func (bot Bot) randomTurnAndInvokeAction() {
	bot.centerView()
	bot.randomTurn(any)
	bot.sleep(800)

	for bot.recognizeAndInvokeAction() {
		bot.randomSleep(actionMinTime, actionMaxTime)
	}
}

func (bot Bot) centerView() {
	robotgo.MoveMouse(500, 600) // todo
}

func main() {
	//todo find window geometry more accurate
	//todo implement harvest area fn

	rand.Seed(time.Now().UnixNano())
	testActionFlag := flag.String("t", "", "test action")
	flag.Parse()

	bot := Bot{getConfiguration()}
	robotgo.ActivePID(bot.config.pid)

	runTestCommandIfFlagSet(bot, *testActionFlag)

	// 1st method of resource gathering (preferred)
	bot.alwaysRunAndInvokeAction()

	for {
		// 2nd and 3d method of resource gathering (based on image recognition)
		//bot.randomTurnAndInvokeAction()
		//bot.randomRunAndInvokeAction()
	}
}
