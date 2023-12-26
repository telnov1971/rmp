package main

import (
	"fmt"
	"time"
)

func scheduleStart() {
	// Создаем расписание с указанными временами
	schedule := []string{"07:00", "12:00", "18:00"}

	// Получаем текущую локальную временную зону
	localZone, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Println("Ошибка при загрузке локальной временной зоны:", err)
		return
	}

	for _, t := range schedule {
		// Парсим время из расписания
		targetTime, err := time.ParseInLocation("15:04", t, localZone)
		if err != nil {
			fmt.Println("Ошибка при парсинге времени:", err)
			return
		}

		// Получаем текущее время
		currentTime := time.Now()

		// Вычисляем время до следующего запуска задачи
		duration := targetTime.Sub(currentTime)

		if duration < 0 {
			// Если время уже прошло на сегодня, переходим к следующему
			continue
		}

		fmt.Println("Задача будет запущена через", duration)

		// Ожидаем до времени запуска задачи
		time.Sleep(duration)

		// Здесь можно вызвать функцию или выполнить нужную задачу
		fmt.Println("Задача запущена в", time.Now())
	}
}
