package main

import (
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Parking struct {
	freeSpotsRand     int
	occupiedSpotsRand int
	autoMode          chan string
	operatingMode     bool
}

func (p *Parking) UpdateSpots() {
	p.freeSpotsRand = rand.Intn(20) + 10
	p.occupiedSpotsRand = 95 - p.freeSpotsRand
}

func mainwindow() {
	a := app.New()
	w := a.NewWindow("Smart Parking")

	//Так как истёк срок
	initMQTTClient("tls://d24e3221.ala.eu-central-1.emqxsl.com:8883", "vadlap", "12345", true, "emqxsl-ca.crt")
	initMQTTClient("tls://dev.rightech.io:8883", "vadlap", "12345", false, "")
	myParking := &Parking{operatingMode: true}

	// Создаем канал для получения данных из подписки
	messageChannel := make(chan string)
	myParking.autoMode = messageChannel

	// Запускаем подписку в отдельной горутине
	go subscribeToTopic("mode/topic", messageChannel)

	// Создаем виджеты для отображения свободных и занятых мест
	freeSpots := widget.NewLabel("Свободные места: 0")
	occupiedSpots := widget.NewLabel("Занятые места: 0")

	// Таймер для обновления значений каждые 5 секунд
	ticker := time.NewTicker(5 * time.Second)

	// Горутина для обновления значений через каждые 5 секунд

	go func() {
		for {
			if myParking.operatingMode {
				<-ticker.C              // Ожидание тикера
				myParking.UpdateSpots() // Обновление данных парковки
				freeSpots.SetText("Свободные места: " + strconv.Itoa(myParking.freeSpotsRand))
				occupiedSpots.SetText("Занятые места: " + strconv.Itoa(myParking.occupiedSpotsRand))
				// sendData(myParking.freeSpotsRand, myParking.occupiedSpotsRand, "vadlap/topic")
				sendData(myParking.freeSpotsRand, "state/freeSpots")
				sendData(myParking.occupiedSpotsRand, "state/occupiedSpots")
			}
		}
	}()

	// Верх: Состояние парковки
	parkingStatus := container.NewGridWithColumns(2, freeSpots, occupiedSpots)

	// Левая нижняя панель: ручной режим, шлагбаум и табло
	manualMode := widget.NewLabel("Ручной режим включён")

	var barrierClosed *widget.Check
	var signOff *widget.Check

	// Инициализация и установка функций для изменения текста
	barrierClosed = widget.NewCheck("Шлагбаум открыт", func(value bool) {
		if value {
			barrierClosed.SetText("Шлагбаум закрыт")
		} else {
			barrierClosed.SetText("Шлагбаум открыт")
		}
	})
	signOff = widget.NewCheck("Табло включено", func(value bool) {
		if value {
			signOff.SetText("Табло выключено")
		} else {
			signOff.SetText("Табло включено")
		}
	})
	leftPanel := container.NewVBox(manualMode, barrierClosed, signOff)

	// Кнопки управления парковкой (в центре)
	openButton := widget.NewButton("Открыть", func() {})
	closeButton := widget.NewButton("Закрыть", func() {})
	parkingControl := container.NewHBox(
		layout.NewSpacer(), // Spacer слева
		openButton,
		closeButton,
		layout.NewSpacer(), // Spacer справа
	)

	// Кнопки выбора режима работы
	manualModeButton := widget.NewButton("Ручной", func() {})
	autoModeButton := widget.NewButton("Автоматический", func() {})
	modeControl := container.NewHBox(
		layout.NewSpacer(), // Spacer слева
		manualModeButton,
		autoModeButton,
		layout.NewSpacer(), // Spacer справа
	)

	manualModeButton.OnTapped = func() {
		if !leftPanel.Visible() {
			leftPanel.Show()
		}
	}

	autoModeButton.OnTapped = func() {
		if leftPanel.Visible() {
			leftPanel.Hide()
		}
	}

	openButton.OnTapped = func() {
		if !freeSpots.Visible() {
			freeSpots.Show()
			occupiedSpots.Show()
		}
		myParking.operatingMode = true
	}

	closeButton.OnTapped = func() {
		if freeSpots.Visible() {
			freeSpots.Hide()
			occupiedSpots.Hide()
		}
		myParking.operatingMode = false
	}

	go func() {
		for mode := range myParking.autoMode {
			if mode == "hand" && !leftPanel.Visible() {
				leftPanel.Show()
			} else if mode == "auto" && leftPanel.Visible() {
				leftPanel.Hide()
			}
		}
	}()

	go func() {
		for mode := range myParking.autoMode {
			if mode == "true" && freeSpots.Visible() {
				freeSpots.Hide()
				occupiedSpots.Hide()
			}
		}
	}()

	// Центральная панель с кнопками
	controlPanel := container.NewVBox(
		layout.NewSpacer(), // Spacer сверху для центрирования
		widget.NewLabel("Работа парковки"),
		parkingControl,
		widget.NewLabel("Режим работы"),
		modeControl,
		layout.NewSpacer(), // Spacer снизу
	)

	// Основной макет с использованием BorderLayout
	content := container.New(
		layout.NewBorderLayout(parkingStatus, nil, leftPanel, nil),
		parkingStatus, // верхняя часть
		leftPanel,     // левая часть (панель снизу)
		controlPanel,  // центральная панель
	)

	// Установка содержимого окна
	w.SetContent(content)
	w.Resize(fyne.NewSize(650, 400))
	w.SetFixedSize(true)
	w.ShowAndRun()
}
