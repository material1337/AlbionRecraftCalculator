package main

import (
	"fmt"
	"image/color"
	"math"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Константы для процентных ставок
const (
	defaultReturnRate           = 0.152 // 15.2%
	defaultFocusRate            = 0.435 // 43.5%
	specialReturnRate           = 0.367 // 36.7%
	specialFocusRate            = 0.539 // 53.9%
	defaultBonusReturnRate      = 0.167 // 16.72% округленные
	defaultBonusFocusReturnRate = 0.479 // 47.85% округленные
	specialBonusReturnRate      = 0.404 // 40.37% округленные
	spacialBonusFocusReturnRate = 0.593 // 59.29% округленные
)

// CustomDarkTheme - наша кастомная темная тема
type CustomDarkTheme struct{}

func (CustomDarkTheme) Color(c fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch c {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	default:
		return theme.DarkTheme().Color(c, v)
	}
}

func (CustomDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}

func (CustomDarkTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(n)
}

func (CustomDarkTheme) Size(s fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(s)
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(CustomDarkTheme{})
	myWindow := myApp.NewWindow("AlbionTools")

	customContainer1 := container.NewWithoutLayout()
	customContainer2 := container.NewWithoutLayout()

	// Объявляем виджеты выбора города и ресурса
	options2 := []string{"Город", "Lymhyrst", "Brightwatch", "Martlock", "Fort Sterling", "Thetford"}
	Town := widget.NewSelect(options2, func(selected string) {})
	Town.SetSelected("Город")
	Town.Move(fyne.NewPos(230, 30))
	Town.Resize(fyne.NewSize(100, 40))

	options3 := []string{"Ресурс", "Брус", "Ткань", "Кожа", "Железо"}
	Resource := widget.NewSelect(options3, func(selected string) {})
	Resource.SetSelected("Ресурс")
	Resource.Move(fyne.NewPos(230, 80))
	Resource.Resize(fyne.NewSize(100, 40))

	var (
		primaryPrice   float64
		secondaryPrice float64
		primaryQty     float64
		secondaryQty   float64
		foodPrice      float64
		craftingPrice  float64
		focusEnabled   bool
		taxRate        float64 = 0.0
		stationCost    float64
		hasBonus       bool = false
	)

	isSpecialCombination := func(resource, town string) bool {
		specialCombinations := map[string]string{
			"Брус":   "Fort Sterling",
			"Ткань":  "Lymhyrst",
			"Кожа":   "Martlock",
			"Железо": "Thetford",
		}

		if resource == "" || town == "" || resource == "Ресурс" || town == "Город" {
			return false
		}

		expectedTown, exists := specialCombinations[resource]
		return exists && town == expectedTown
	}

	calculateRecursive := func(initialAmount float64) float64 {
		var returnRate float64
		isSpecial := isSpecialCombination(Resource.Selected, Town.Selected)

		switch {
		case isSpecial && focusEnabled && hasBonus:
			returnRate = spacialBonusFocusReturnRate
		case isSpecial && focusEnabled:
			returnRate = specialFocusRate
		case isSpecial && hasBonus:
			returnRate = specialBonusReturnRate
		case isSpecial:
			returnRate = specialReturnRate
		case focusEnabled && hasBonus:
			returnRate = defaultBonusFocusReturnRate
		case focusEnabled:
			returnRate = defaultFocusRate
		case hasBonus:
			returnRate = defaultBonusReturnRate
		default:
			returnRate = defaultReturnRate
		}

		total := initialAmount
		currentReturn := initialAmount * returnRate

		for currentReturn >= 1 {
			total += currentReturn
			currentReturn = currentReturn * returnRate
		}

		return total
	}

	// Виджеты результатов
	kolvo := widget.NewEntry()
	kolvo.Move(fyne.NewPos(340, 180))
	kolvo.Resize(fyne.NewSize(100, 40))
	kolvo.Disable()

	trati := widget.NewEntry()
	trati.Move(fyne.NewPos(450, 180))
	trati.Resize(fyne.NewSize(100, 40))
	trati.Disable()

	pribil := widget.NewEntry()
	pribil.Move(fyne.NewPos(560, 180))
	pribil.Resize(fyne.NewSize(100, 40))
	pribil.Disable()

	updateCalculations := func() {
		var totalResources float64
		if secondaryQty > 0 {
			totalResources = calculateRecursive(secondaryQty)
			kolvo.SetText(fmt.Sprintf("%.0f", math.Floor(totalResources)))
		} else {
			kolvo.SetText("0")
		}

		// Расчет стоимости станков (умножаем на количество готового)
		stationTotalCost := stationCost * math.Floor(totalResources)

		// Расчет общих затрат (материалы + еда + станки)
		totalCost := primaryPrice*primaryQty + secondaryPrice*secondaryQty + foodPrice + stationTotalCost
		trati.SetText(fmt.Sprintf("%.0f", totalCost))

		if craftingPrice > 0 && totalResources > 0 {
			grossProfit := craftingPrice * math.Floor(totalResources)
			netProfit := grossProfit - totalCost
			profitAfterTax := netProfit * (1 - taxRate)
			pribil.SetText(fmt.Sprintf("%.0f", profitAfterTax))
		} else {
			pribil.SetText("0")
		}
	}

	Town.OnChanged = func(selected string) {
		updateCalculations()
	}
	Resource.OnChanged = func(selected string) {
		updateCalculations()
	}

	primary := widget.NewEntry()
	primary.SetPlaceHolder("Первичка:")
	primary.Move(fyne.NewPos(10, 30))
	primary.Resize(fyne.NewSize(100, 40))
	primary.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			primaryPrice = val
			updateCalculations()
		}
	}

	secondary := widget.NewEntry()
	secondary.SetPlaceHolder("Вторичка:")
	secondary.Move(fyne.NewPos(10, 80))
	secondary.Resize(fyne.NewSize(100, 40))
	secondary.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			secondaryPrice = val
			updateCalculations()
		}
	}

	primaryAmount := widget.NewEntry()
	primaryAmount.SetPlaceHolder("Первичка:")
	primaryAmount.Move(fyne.NewPos(340, 30))
	primaryAmount.Resize(fyne.NewSize(100, 40))
	primaryAmount.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			primaryQty = val
			updateCalculations()
		}
	}

	secondaryAmount := widget.NewEntry()
	secondaryAmount.SetPlaceHolder("Вторичка:")
	secondaryAmount.Move(fyne.NewPos(340, 80))
	secondaryAmount.Resize(fyne.NewSize(100, 40))
	secondaryAmount.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			secondaryQty = val
			updateCalculations()
		}
	}

	complitellyPrice := widget.NewEntry()
	complitellyPrice.SetPlaceHolder("Шт готов:")
	complitellyPrice.Move(fyne.NewPos(10, 130))
	complitellyPrice.Resize(fyne.NewSize(100, 40))
	complitellyPrice.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			craftingPrice = val
			updateCalculations()
		}
	}

	focus := widget.NewCheck("Фокус", func(checked bool) {
		focusEnabled = checked
		updateCalculations()
	})
	focus.Move(fyne.NewPos(120, 30))
	focus.Resize(fyne.NewSize(100, 40))

	options := []string{"Бонус", "10%", "Нэту("}
	bonus := widget.NewSelect(options, func(selected string) {
		hasBonus = selected == "10%"
		updateCalculations()
	})
	bonus.SetSelected("Бонус")
	bonus.Move(fyne.NewPos(120, 80))
	bonus.Resize(fyne.NewSize(100, 40))

	eat := widget.NewEntry()
	eat.SetPlaceHolder("Цена еды:")
	eat.Move(fyne.NewPos(10, 180))
	eat.Resize(fyne.NewSize(100, 40))
	eat.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			foodPrice = val
			updateCalculations()
		}
	}

	options1 := []string{"Налог", "6.5%", "10.5%"}
	buyer := widget.NewSelect(options1, func(selected string) {
		switch selected {
		case "Налог":
			taxRate = 0.0
		case "6.5%":
			taxRate = 0.065
		case "10.5%":
			taxRate = 0.15
		}
		updateCalculations()
	})
	buyer.SetSelected("Налог")
	buyer.Move(fyne.NewPos(120, 130))
	buyer.Resize(fyne.NewSize(100, 40))

	Buyer2 := widget.NewEntry()
	Buyer2.SetPlaceHolder("Налог")
	Buyer2.Move(fyne.NewPos(120, 180))
	Buyer2.Resize(fyne.NewSize(100, 40))
	Buyer2.OnChanged = func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			stationCost = val
			updateCalculations()
		}
	}

	options4 := []string{"Тир", "4", "5", "6", "7", "8"}
	Tier := widget.NewSelect(options4, func(selected string) {})
	Tier.SetSelected("Тир")
	Tier.Move(fyne.NewPos(230, 130))
	Tier.Resize(fyne.NewSize(100, 40))

	image := canvas.NewImageFromResource(resourceRecraft113Png)
	image.FillMode = canvas.ImageFillStretch
	image.Move(fyne.NewPos(450, 10))
	image.Resize(fyne.NewSize(210, 140))

	createStyledLabel := func(text string, pos fyne.Position, size fyne.Size) *canvas.Text {
		label := canvas.NewText(text, color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff})
		label.TextSize = theme.TextSize()
		label.Alignment = fyne.TextAlignCenter
		label.Move(pos)
		label.Resize(size)
		return label
	}

	label1 := createStyledLabel("Цена", fyne.NewPos(10, 0), fyne.NewSize(100, 20))
	label11 := createStyledLabel("Кол-во", fyne.NewPos(340, 0), fyne.NewSize(100, 20))
	label3 := createStyledLabel("Доп данные", fyne.NewPos(120, 0), fyne.NewSize(200, 20))
	label4 := createStyledLabel("Прибыль", fyne.NewPos(560, 150), fyne.NewSize(100, 20))
	label5 := createStyledLabel("Траты", fyne.NewPos(450, 150), fyne.NewSize(100, 20))
	label6 := createStyledLabel("Кол-во готового", fyne.NewPos(335, 150), fyne.NewSize(100, 20))
	label7 := createStyledLabel("*Станка за шт*", fyne.NewPos(120, 220), fyne.NewSize(100, 20))

	// Вторая вкладка
	label10 := createStyledLabel("Material1337", fyne.NewPos(250, 160), fyne.NewSize(200, 20))

	link1, err := url.Parse("https://github.com/material1337")
	if err != nil {
		panic(err)
	}

	hyperlink1 := widget.NewHyperlink("GitHub", link1)
	hyperlink1.Move(fyne.NewPos(240, 190))
	hyperlink1.Resize(fyne.NewSize(200, 20))
	hyperlink1.Alignment = fyne.TextAlignCenter

	link2, err := url.Parse("https://t.me/Material1337")
	if err != nil {
		panic(err)
	}

	hyperlink2 := widget.NewHyperlink("Telegram", link2)
	hyperlink2.Move(fyne.NewPos(240, 220))
	hyperlink2.Resize(fyne.NewSize(200, 20))
	hyperlink2.Alignment = fyne.TextAlignCenter

	image2 := canvas.NewImageFromResource(resourceAVaAlbPng)
	image2.FillMode = canvas.ImageFillStretch
	image2.Move(fyne.NewPos(240, 10))
	image2.Resize(fyne.NewSize(200, 150))

	// Добавление виджетов в контейнеры
	customContainer1.Add(primary)
	customContainer1.Add(secondary)
	customContainer1.Add(primaryAmount)
	customContainer1.Add(secondaryAmount)
	customContainer1.Add(complitellyPrice)
	customContainer1.Add(focus)
	customContainer1.Add(bonus)
	customContainer1.Add(eat)
	customContainer1.Add(buyer)
	customContainer1.Add(kolvo)
	customContainer1.Add(pribil)
	customContainer1.Add(trati)
	customContainer1.Add(Town)
	customContainer1.Add(Resource)
	customContainer1.Add(Tier)
	customContainer1.Add(label1)
	customContainer1.Add(label11)
	customContainer1.Add(label3)
	customContainer1.Add(label4)
	customContainer1.Add(label5)
	customContainer1.Add(label6)
	customContainer1.Add(image)
	customContainer1.Add(Buyer2)
	customContainer1.Add(label7)

	customContainer2.Add(image2)
	customContainer2.Add(hyperlink1)
	customContainer2.Add(hyperlink2)
	customContainer2.Add(label10)

	tabs := container.NewAppTabs(
		container.NewTabItem("Перекрафт", customContainer1),
		container.NewTabItem("Контакты", customContainer2),
	)

	myWindow.SetContent(tabs)
	myWindow.SetFixedSize(true)
	myWindow.Resize(fyne.NewSize(680, 310))
	myWindow.ShowAndRun()
}
