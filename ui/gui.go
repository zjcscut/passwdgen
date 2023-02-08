package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"passwdgen/gen"
	"passwdgen/i18n"
	pm "passwdgen/theme"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTheme    = "Dark(暗黑)"
	DefaultLanguage = "zh_CN(中文)"
)

func InitMainWindow() fyne.Window {
	app := fyne.CurrentApp()
	mainWindow := app.NewWindow("")
	i18n.RegisterRefresher(i18n.MainWindowTitleKey, func(value string) {
		mainWindow.SetTitle(value)
	})
	var canvasObjectsToRefresh []fyne.CanvasObject
	callback := func() []fyne.CanvasObject {
		return canvasObjectsToRefresh
	}
	// 密码生成Tab
	mainTabItem := initMainTabContent(mainWindow)
	canvasObjectsToRefresh = append(canvasObjectsToRefresh, mainTabItem)
	passwdTab := container.NewTabItemWithIcon("", theme.HomeIcon(), mainTabItem)
	i18n.RegisterRefresher(i18n.PassWdTabTitleKey, func(value string) {
		passwdTab.Text = value
	})
	// 设置Tab
	settingTabItem, tls := initSettingTabContent(callback, mainWindow)
	canvasObjectsToRefresh = append(canvasObjectsToRefresh, settingTabItem)
	settingTab := container.NewTabItemWithIcon("", theme.SettingsIcon(), settingTabItem)
	i18n.RegisterRefresher(i18n.SettingTabTitleKey, func(value string) {
		settingTab.Text = value
	})
	tabs := container.NewAppTabs(passwdTab, settingTab)
	// 选中的Tab进行刷新
	tabs.OnSelected = func(ti *container.TabItem) {
		ti.Content.Refresh()
	}
	mainWindow.SetContent(tabs)
	// 选择默认主题和语言
	tls.selectDefault()
	return mainWindow
}

func initMainTabContent(w fyne.Window) fyne.CanvasObject {
	bindings := newDefaultBindings()
	// 随机密码输出
	passwdOutputEntry := widget.NewEntryWithData(bindings.passwdOutputBinding)
	// 密码强度
	passwdStrengthLabel := widget.NewLabel("")
	i18n.RegisterRefresher(i18n.PasswdStrengthLabelKey, func(value string) {
		passwdStrengthLabel.Text = value
	})
	bindings.passwdStrengthInfo = canvas.NewText("", nil)
	bindings.passwdStrengthCost = canvas.NewText("", color.NRGBA{R: 0xff, G: 0x98, B: 0x00, A: 0xff})
	// 密码长度
	passwdLengthLabel := widget.NewLabel("")
	i18n.RegisterRefresher(i18n.PasswdLengthLabelKey, func(value string) {
		passwdLengthLabel.Text = value
	})
	bindings.passwdLengthInfo = canvas.NewText("0", color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff})
	// 监听密码长度变动
	bindings.passwdLengthBinding.AddListener(binding.NewDataListener(func() {
		generatePassword(w, bindings)
	}))
	passwdLengthSlide := widget.NewSliderWithData(0, 64, bindings.passwdLengthBinding)
	passwdLengthSlideLabel := widget.NewLabel("")
	i18n.RegisterRefresher(i18n.PasswdLengthSlideLabelKey, func(value string) {
		passwdLengthSlideLabel.Text = value
	})
	passwdLengthSlide.Step = 1
	passwdLengthSlide.OnChanged = func(value float64) {
		_ = bindings.passwdLengthBinding.Set(value)
		bindings.passwdLengthInfo.Text = fmt.Sprintf("%0.0f", value)
		bindings.passwdLengthInfo.Refresh()
	}
	passwdLengthButtons := container.NewGridWithColumns(4,
		widget.NewButton("8", func() {
			_ = bindings.passwdLengthBinding.Set(8)
		}),
		widget.NewButton("16", func() {
			_ = bindings.passwdLengthBinding.Set(16)
		}),
		widget.NewButton("32", func() {
			_ = bindings.passwdLengthBinding.Set(32)
		}),
		widget.NewButton("64", func() {
			_ = bindings.passwdLengthBinding.Set(64)
		}))
	pslc := container.NewGridWithColumns(2,
		container.NewHBox(passwdStrengthLabel, bindings.passwdStrengthInfo, bindings.passwdStrengthCost),
		container.New(layout.NewFormLayout(), passwdLengthLabel, bindings.passwdLengthInfo),
	)
	plc := container.NewGridWithColumns(2, container.New(layout.NewFormLayout(), passwdLengthSlideLabel,
		passwdLengthSlide), passwdLengthButtons)
	// 选项
	numberCheck := newCheckWidget("", i18n.NumberCheckLabelKey, func(check bool) {
		_ = bindings.enableNumber.Set(check)
		generatePassword(w, bindings)
	}, getBoolBindingValue(bindings.enableNumber))
	lowercaseCheck := newCheckWidget("", i18n.LowercaseCheckLabelKey, func(check bool) {
		_ = bindings.enableLowercase.Set(check)
		generatePassword(w, bindings)
	}, getBoolBindingValue(bindings.enableLowercase))
	uppercaseCheck := newCheckWidget("", i18n.UppercaseCheckLabelKey, func(check bool) {
		_ = bindings.enableUppercase.Set(check)
		generatePassword(w, bindings)
	}, getBoolBindingValue(bindings.enableUppercase))
	duplicateCheck := newCheckWidget("", i18n.DuplicateCheckLabelKey, func(check bool) {
		_ = bindings.enableDuplicate.Set(check)
		generatePassword(w, bindings)
	}, getBoolBindingValue(bindings.enableDuplicate))
	checkGroup := container.New(layout.NewGridLayout(4), numberCheck, lowercaseCheck, uppercaseCheck, duplicateCheck)
	// 包含特殊字符
	includeSpecialCharSetForm := newCharSetEntryContainer("", i18n.IncludeSpecialCharSetFormLabelKey, bindings.includeSpecialCharSet)
	// 排除特殊字符
	excludeSpecialCharSetForm := newCharSetEntryContainer("", i18n.ExcludeSpecialCharSetFormLabelKey, bindings.excludeSpecialCharSet)
	// 功能按钮组
	copyButton := newOptionButtonWidget("", i18n.CopyButtonLabelKey, theme.ContentCopyIcon(), func() {
		value, _ := bindings.passwdOutputBinding.Get()
		w.Clipboard().SetContent(value)
	})
	generateButton := newOptionButtonWidget("", i18n.GenerateButtonLabelKey, theme.NavigateNextIcon(), func() {
		generatePassword(w, bindings)
	})
	resetButton := newOptionButtonWidget("", i18n.ResetButtonLabelKey, theme.ViewRefreshIcon(), func() {
		application := fyne.CurrentApp()
		application.Preferences().SetBool("__Resetting__", true)
		resetBindings(bindings)
		numberCheck.Checked = getBoolBindingValue(bindings.enableNumber)
		numberCheck.Refresh()
		lowercaseCheck.Checked = getBoolBindingValue(bindings.enableLowercase)
		lowercaseCheck.Refresh()
		uppercaseCheck.Checked = getBoolBindingValue(bindings.enableUppercase)
		uppercaseCheck.Refresh()
		duplicateCheck.Checked = getBoolBindingValue(bindings.enableDuplicate)
		duplicateCheck.Refresh()
		application.Preferences().SetBool("__Resetting__", false)
	})
	optionButtonGroup := container.New(layout.NewGridLayout(3), copyButton, generateButton, resetButton)
	passwdGenBox := container.NewVBox(
		passwdOutputEntry,
		pslc,
		plc,
		checkGroup,
		includeSpecialCharSetForm,
		excludeSpecialCharSetForm,
		optionButtonGroup,
	)
	passwdGenCard := widget.NewCard("", "", passwdGenBox)
	i18n.RegisterRefresher(i18n.PasswdGenCardTitleKey, func(value string) {
		passwdGenCard.Title = value
	})
	passwdGenBorder := container.NewBorder(passwdGenCard, nil, nil, nil)
	var historyRecordSlice []*historyRecordItem
	// 历史记录
	historyRecordTable := widget.NewTable(func() (int, int) {
		return len(historyRecordSlice), 4
	}, func() fyne.CanvasObject {
		return container.NewMax(
			widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}),
			widget.NewToolbar(
				widget.NewToolbarAction(theme.ContentCopyIcon(), func() {

				}),
				widget.NewToolbarSeparator(),
				widget.NewToolbarAction(theme.DeleteIcon(), func() {

				}),
			),
		)
	}, func(cellId widget.TableCellID, object fyne.CanvasObject) {

	})
	historyRecordTable.UpdateCell = func(cellId widget.TableCellID, object fyne.CanvasObject) {
		historyRecordLabel := object.(*fyne.Container).Objects[0].(*widget.Label)
		historyRecordToolbar := object.(*fyne.Container).Objects[1].(*widget.Toolbar)
		historyRecordLabel.Show()
		historyRecordToolbar.Hide()
		row := cellId.Row
		switch cellId.Col {
		case 0:
			// 序号
			historyRecordLabel.SetText(strconv.Itoa(row + 1))
		case 1:
			// 密码
			l := len(historyRecordSlice)
			if l > 0 && row < l {
				historyItem := historyRecordSlice[row]
				if historyItem != nil {
					historyRecordLabel.SetText(historyItem.password)
				}
			}
		case 2:
			// 日期
			l := len(historyRecordSlice)
			if l > 0 && row < l {
				historyItem := historyRecordSlice[row]
				if historyItem != nil {
					historyRecordLabel.SetText(historyItem.createTime)
				}
			}
		case 3:
			// 拷贝\删除按钮
			historyRecordLabel.Hide()
			historyRecordToolbar.Show()
			historyRecordToolbarItems := historyRecordToolbar.Items
			historyRecordToolbarItems[0] = widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
				l := len(historyRecordSlice)
				if l > 0 && row < l {
					historyItem := historyRecordSlice[row]
					if historyItem != nil && historyItem.password != "" {
						w.Clipboard().SetContent(historyItem.password)
					}
				}
			})
			historyRecordToolbarItems[2] = widget.NewToolbarAction(theme.DeleteIcon(), func() {
				l := len(historyRecordSlice)
				if l > 0 && row < l {
					historyRecordSlice = removeHistoryItemByIndex(historyRecordSlice, row)
					historyRecordTable.Refresh()
				}
			})
			historyRecordToolbar.Refresh()
		}
	}
	historyRecordTable.SetColumnWidth(0, 32)
	historyRecordTable.SetColumnWidth(1, 200)
	historyRecordTable.SetColumnWidth(2, 160)
	historyRecordScroll := container.NewVScroll(historyRecordTable)
	historyRecordScroll.SetMinSize(fyne.NewSize(0, 200))
	historyRecordTable.Hide()
	historyCheck := newCheckWidget("", i18n.HistoryCheckLabelKey, func(check bool) {
		if !check {
			historyRecordScroll.Hide()
			historyRecordTable.Hide()
			historyRecordSlice = nil
			historyRecordTable.Refresh()
			historyRecordScroll.Refresh()
		} else {
			historyRecordScroll.Show()
			historyRecordTable.Show()
		}
	}, false)
	historyBox := container.NewVBox(
		historyCheck,
		historyRecordScroll,
	)
	historyCard := widget.NewCard("", "", historyBox)
	i18n.RegisterRefresher(i18n.HistoryCardTitleKey, func(value string) {
		historyCard.Title = value
	})
	historyBorder := container.NewBorder(historyCard, nil, nil, nil)
	go func() {
		for {
			select {
			case hri := <-bindings.historyRecordChan:
				if historyCheck.Checked {
					historyRecordSlice = append(historyRecordSlice, hri)
					historyRecordTable.Refresh()
				}
			}
		}
	}()
	return container.NewVBox(passwdGenBorder, historyBorder)
}

func initSettingTabContent(callback func() []fyne.CanvasObject, _ fyne.Window) (fyne.CanvasObject, *themeLangSelector) {
	app := fyne.CurrentApp()
	themeGroup := widget.NewRadioGroup([]string{
		"Dark(暗黑)",
		"Light(白色)",
	}, func(st string) {
		app.Settings().SetTheme(&pm.SelectableTheme{Theme: st})
	})
	themeGroup.Required = true
	themeGroup.Horizontal = true
	langGroup := widget.NewRadioGroup([]string{
		"zh_CN(中文)",
		"en_US(英文)",
	}, func(lang string) {
		canvasLocalizer := &i18n.CanvasLocalizer{
			Lang: lang,
		}
		canvasLocalizer.SwitchLang(callback())
	})
	langGroup.Horizontal = true
	langGroup.Required = true
	themeForm := widget.NewFormItem("", themeGroup)
	i18n.RegisterRefresher(i18n.SettingThemeFormTitleKey, func(value string) {
		themeForm.Text = value
	})
	langForm := widget.NewFormItem("", langGroup)
	i18n.RegisterRefresher(i18n.SettingLangFormTitleKey, func(value string) {
		langForm.Text = value
	})
	uiForm := container.NewVBox(widget.NewForm(themeForm, langForm))
	appearanceCard := widget.NewCard("", "", uiForm)
	i18n.RegisterRefresher(i18n.SettingAppearanceCardTitleKey, func(value string) {
		appearanceCard.Title = value
	})
	box := container.NewVBox(appearanceCard)
	return container.NewBorder(box, nil, nil, nil), &themeLangSelector{themeGroup: themeGroup,
		langGroup: langGroup}
}

func newCharSetEntryContainer(label string, i18nKey i18n.MessageId, dataBinding binding.String) *fyne.Container {
	labelWidget := widget.NewLabel(label)
	i18n.RegisterRefresher(i18nKey, func(value string) {
		labelWidget.Text = value
	})
	entry := widget.NewEntryWithData(dataBinding)
	return container.New(layout.NewFormLayout(), labelWidget, entry)
}

func newOptionButtonWidget(label string, i18nKey i18n.MessageId, icon fyne.Resource, callback func()) *widget.Button {
	button := widget.NewButtonWithIcon(label, icon, callback)
	i18n.RegisterRefresher(i18nKey, func(value string) {
		button.Text = value
	})
	return button
}

func newCheckWidget(label string, i18nKey i18n.MessageId, callback func(check bool), checked bool) *widget.Check {
	check := widget.NewCheck(label, callback)
	i18n.RegisterRefresher(i18nKey, func(value string) {
		check.Text = value
	})
	check.Checked = checked
	return check
}

func getBoolBindingValue(b binding.Bool) bool {
	r, _ := b.Get()
	return r
}

func getUint8FromFloat64BindingValue(f binding.Float) uint8 {
	r, _ := f.Get()
	return uint8(r)
}

func getStringBindingValue(s binding.String) string {
	r, err := s.Get()
	if err != nil {
		return ""
	}
	return r
}

type bindings struct {
	// 生成密码输出的数据绑定
	passwdOutputBinding binding.String
	// 密码强度描述
	passwdStrengthInfo *canvas.Text
	// 密码强度破解耗时
	passwdStrengthCost *canvas.Text
	// 密码长度描述
	passwdLengthInfo *canvas.Text
	// 密码长度数据绑定
	passwdLengthBinding binding.Float
	// 是否使用数字
	enableNumber binding.Bool
	// 是否使用小写字母
	enableLowercase binding.Bool
	// 是否使用大写字母
	enableUppercase binding.Bool
	// 是否允许重复
	enableDuplicate binding.Bool
	// 包含的特殊字符集
	includeSpecialCharSet binding.String
	// 排除的特殊字符集
	excludeSpecialCharSet binding.String
	// 历史记录Channel
	historyRecordChan chan *historyRecordItem
}

type historyRecordItem struct {
	password   string
	createTime string
}

func newDefaultBindings() *bindings {
	defaultConf := gen.NewDefaultPasswdGenConf()
	bindings := &bindings{
		passwdOutputBinding:   binding.NewString(),
		passwdLengthBinding:   binding.NewFloat(),
		enableNumber:          binding.NewBool(),
		enableLowercase:       binding.NewBool(),
		enableUppercase:       binding.NewBool(),
		enableDuplicate:       binding.NewBool(),
		includeSpecialCharSet: binding.NewString(),
		excludeSpecialCharSet: binding.NewString(),
		historyRecordChan:     make(chan *historyRecordItem),
	}
	_ = bindings.passwdLengthBinding.Set(float64(defaultConf.Length))
	_ = bindings.enableNumber.Set(defaultConf.EnableNumber)
	_ = bindings.enableLowercase.Set(defaultConf.EnableLowercase)
	_ = bindings.enableUppercase.Set(defaultConf.EnableUppercase)
	_ = bindings.enableDuplicate.Set(defaultConf.EnableDuplicate)
	_ = bindings.includeSpecialCharSet.Set(defaultConf.IncludeSpecialCharSet)
	_ = bindings.excludeSpecialCharSet.Set(defaultConf.ExcludeSpecialCharSet)
	return bindings
}

func resetBindings(bindings *bindings) {
	defaultConf := gen.NewDefaultPasswdGenConf()
	_ = bindings.passwdLengthBinding.Set(float64(defaultConf.Length))
	_ = bindings.passwdOutputBinding.Set("")
	_ = bindings.enableNumber.Set(defaultConf.EnableNumber)
	_ = bindings.enableLowercase.Set(defaultConf.EnableLowercase)
	_ = bindings.enableUppercase.Set(defaultConf.EnableUppercase)
	_ = bindings.enableDuplicate.Set(defaultConf.EnableDuplicate)
	_ = bindings.includeSpecialCharSet.Set(defaultConf.IncludeSpecialCharSet)
	_ = bindings.excludeSpecialCharSet.Set(defaultConf.ExcludeSpecialCharSet)
	if bindings.passwdStrengthInfo != nil {
		bindings.passwdStrengthInfo.Text = ""
	}
	if bindings.passwdStrengthCost != nil {
		bindings.passwdStrengthCost.Text = ""
	}
}

func generatePassword(w fyne.Window, bindings *bindings) {
	application := fyne.CurrentApp()
	// 主窗口初始化才处理密码生成
	if application.Preferences().Bool("__MainWindowInit__") && !application.Preferences().Bool("__Resetting__") {
		result, err := gen.GeneratePassword(&gen.PasswdGenConf{
			Length:                getUint8FromFloat64BindingValue(bindings.passwdLengthBinding),
			EnableNumber:          getBoolBindingValue(bindings.enableNumber),
			EnableLowercase:       getBoolBindingValue(bindings.enableLowercase),
			EnableUppercase:       getBoolBindingValue(bindings.enableUppercase),
			EnableDuplicate:       getBoolBindingValue(bindings.enableDuplicate),
			IncludeSpecialCharSet: getStringBindingValue(bindings.includeSpecialCharSet),
			ExcludeSpecialCharSet: getStringBindingValue(bindings.excludeSpecialCharSet),
		})
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			_ = bindings.passwdOutputBinding.Set(result.Password)
			bindings.historyRecordChan <- &historyRecordItem{password: result.Password, createTime: time.Now().Format("2006-01-02 15:04:05")}
			bindings.passwdStrengthInfo.Text = result.StrengthInfo
			bindings.passwdStrengthInfo.Color = result.StrengthColor
			bindings.passwdStrengthInfo.Refresh()
			// CostInfo => StrengthInt
			bindings.passwdStrengthCost.Text = strings.Join([]string{result.CostInfo, " => ",
				strconv.FormatFloat(result.StrengthInt, 'f', 4, 64)}, "")
			bindings.passwdStrengthCost.Color = result.CostColor
			bindings.passwdStrengthCost.Refresh()
		}
	}
}

type themeLangSelector struct {
	themeGroup *widget.RadioGroup
	langGroup  *widget.RadioGroup
}

func (tls *themeLangSelector) selectDefault() {
	tls.themeGroup.SetSelected(DefaultTheme)
	tls.langGroup.SetSelected(DefaultLanguage)
}

func removeHistoryItemByIndex(slice []*historyRecordItem, i int) []*historyRecordItem {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}
