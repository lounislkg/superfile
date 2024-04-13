package components

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func SideBarRender(m model) string {
	s := sidebarTitle.Render(" Super Files")
	s += "\n\n"

	// Ugly shit workaround from hell code, made by @lescx
	amountWellKnownDirectories := len(getWellKnownDirectories())
	amountPinnedDirectories := len(getPinnedDirectories())
	pinnedRendered := false
	externalRendered := false

	for i, directory := range m.sidebarModel.directories {
		if i == amountWellKnownDirectories && !pinnedRendered {
			s += "\n" + sidebarTitle.Render("󰐃 Pinned") + borderStyle.Render(" ───────────") + "\n\n"
			pinnedRendered = true
		}
		if i == amountPinnedDirectories+amountWellKnownDirectories && !externalRendered {
			s += "\n" + sidebarTitle.Render("󱇰 Disks") + borderStyle.Render(" ────────────") + "\n\n"
			externalRendered = true
		}
		cursor := " "
		if m.sidebarModel.cursor == i && m.focusPanel == sidebarFocus {
			cursor = ""
		}
		if directory.location == m.fileModel.filePanels[m.filePanelFocusIndex].location {
			s += cursorStyle.Render(cursor) + sidebarSelected.Render(" "+TruncateText(directory.name, sidebarWidth-2)) + "\n"
		} else {
			s += cursorStyle.Render(cursor) + sidebarItem.Render(" "+TruncateText(directory.name, sidebarWidth-2)) + "\n"
		}
	}

	// In case no pinned directories or external drives are pinned,
	// list menu item at the bottom
	if !pinnedRendered {
		s += "\n" + sidebarTitle.Render("󰐃 Pinned") + borderStyle.Render(" ───────────") + "\n\n"
	}
	if !externalRendered {
		s += "\n" + sidebarTitle.Render("󱇰 Disks") + borderStyle.Render(" ────────────") + "\n\n"
	}

	return SideBarBoardStyle(m.mainPanelHeight, m.focusPanel).Render(s)
}

func FilePanelRender(m model) string {
	// file panel
	f := make([]string, 4)
	for i, filePanel := range m.fileModel.filePanels {
		fileElenent := returnFolderElement(filePanel.location, m.toggleDotFile)
		filePanel.element = fileElenent
		m.fileModel.filePanels[i].element = fileElenent

		f[i] += filePanelTopFolderIcon.Render("   ") + filePanelTopPath.Render(TruncateTextBeginning(filePanel.location, m.fileModel.width-4)) + "\n"
		filePanelWidth := 0
		bottomBorderWidth := 0
		if (m.fullWidth-sidebarWidth-(4+(len(m.fileModel.filePanels)-1)*2))%len(m.fileModel.filePanels) != 0 && i == len(m.fileModel.filePanels)-1 {
			filePanelWidth = (m.fileModel.width + 1)
			bottomBorderWidth = m.fileModel.width + 7
		} else {
			filePanelWidth = m.fileModel.width
			bottomBorderWidth = m.fileModel.width + 6
		}
		f[i] += FilePanelDividerStyle(filePanel.focusType).Render(strings.Repeat("━", filePanelWidth)) + "\n"
		if len(filePanel.element) == 0 {
			f[i] += "   No such file or directory"
			bottomBorder := GenerateBottomBorder("0/0", m.fileModel.width+5)
			f[i] = FilePanelBoardStyle(m.mainPanelHeight, m.fileModel.width, filePanel.focusType, bottomBorder).Render(f[i])
		} else {
			for h := filePanel.render; h < filePanel.render+panelElementHeight(m.mainPanelHeight) && h < len(filePanel.element); h++ {
				cursor := " "
				if h == filePanel.cursor {
					cursor = ""
				}
				isItemSelected := arrayContains(filePanel.selected, filePanel.element[h].location)
				if filePanel.renaming && h == filePanel.cursor {
					f[i] += filePanel.rename.View() + "\n"
				} else {
					f[i] += cursorStyle.Render(cursor+" ") + PrettierName(filePanel.element[h].name, m.fileModel.width-5, filePanel.element[h].directory, isItemSelected) + "\n"
				}
			}
			cursorPosition := strconv.Itoa(filePanel.cursor + 1)
			totalElement := strconv.Itoa(len(filePanel.element))
			panelModeString := ""
			if filePanel.panelMode == browserMode {
				panelModeString = "󰈈 Browser"
			} else if filePanel.panelMode == selectMode {
				panelModeString = "󰆽 Select"
			}
			bottomBorder := GenerateBottomBorder(fmt.Sprintf("%s┣━┫%s/%s", panelModeString, cursorPosition, totalElement), bottomBorderWidth)
			f[i] = FilePanelBoardStyle(m.mainPanelHeight, filePanelWidth, filePanel.focusType, bottomBorder).Render(f[i])
		}
	}

	// file panel render togther
	filePanelRender := ""
	for _, f := range f {
		filePanelRender = lipgloss.JoinHorizontal(lipgloss.Top, filePanelRender, f)
	}
	return filePanelRender
}

func ProcessBarRender(m model) string {
	// save process in the array
	var processes []process
	for _, p := range m.processBarModel.process {
		processes = append(processes, p)
	}

	// sort by the process
	sort.Slice(processes, func(i, j int) bool {
		doneI := (processes[i].state == successful)
		doneJ := (processes[j].state == successful)

		// sort by done or not
		if doneI != doneJ {
			return !doneI
		}

		// if both not done
		if !doneI {
			completionI := float64(processes[i].done) / float64(processes[i].total)
			completionJ := float64(processes[j].done) / float64(processes[j].total)
			return completionI < completionJ // Those who finish first will be ranked later.
		}

		// if both done sort by the doneTime
		return processes[j].doneTime.Before(processes[i].doneTime)
	})

	// render
	processRender := ""
	renderTimes := 0

	for i := m.processBarModel.render; i < len(processes); i++ {
		if renderTimes == 3 {
			break
		}
		process := processes[i]
		process.progress.Width = BottomWidth(m.fullWidth) - 3
		symbol := ""
		cursor := ""
		if i == m.processBarModel.cursor {
			cursor = StringColorRender(theme.Cursor).Render("┃ ")
		} else {
			cursor = StringColorRender(theme.Cursor).Render("  ")
		}
		switch process.state {
		case failure:
			symbol = StringColorRender(theme.Fail).Render("")
		case successful:
			symbol = StringColorRender(theme.Done).Render("")
		case inOperation:
			symbol = StringColorRender(theme.InOperation).Render("󰥔")
		case cancel:
			symbol = StringColorRender(theme.Cancel).Render("")
		}

		processRender += cursor + textStyle.Render(TruncateText(process.name, BottomWidth(m.fullWidth)-7)+" ") + symbol + "\n"
		if renderTimes == 2 {
			processRender += cursor + process.progress.ViewAs(float64(process.done)/float64(process.total)) + ""
		} else {
			processRender += cursor + process.progress.ViewAs(float64(process.done)/float64(process.total)) + "\n\n"
		}
		renderTimes++
	}

	if len(processes) == 0 {
		processRender += "\n   No processes running"
	}
	courseNumber := 0
	if len(m.processBarModel.processList) == 0 {
		courseNumber = 0
	} else {
		courseNumber = m.processBarModel.cursor + 1
	}
	bottomBorder := GenerateBottomBorder(fmt.Sprintf("%s/%s", strconv.Itoa(courseNumber), strconv.Itoa(len(m.processBarModel.processList))), BottomWidth(m.fullWidth)-3)
	processRender = ProcsssBarBoarder(bottomElementHight(footerHeight), BottomWidth(m.fullWidth), bottomBorder, m.focusPanel).Render(processRender)

	return processRender
}

func MetaDataRender(m model) string {
	// process bar
	metaDataBar := ""
	if len(m.fileMetaData.metaData) == 0 && len(m.fileModel.filePanels[m.filePanelFocusIndex].element) > 0 && !m.fileModel.renaming {
		m.fileMetaData.metaData = append(m.fileMetaData.metaData, [2]string{"", ""})
		m.fileMetaData.metaData = append(m.fileMetaData.metaData, [2]string{" 󰥔  Loading metadata...", ""})
		go func() {
			m = returnMetaData(m)
		}()
	}
	maxKeyLength := 0
	sort.Slice(m.fileMetaData.metaData, func(i, j int) bool {
		comparisonFields := []string{"FileName", "FileSize", "FolderName", "FolderSize", "FileModifyDate", "FileAccessDate"}

		for _, field := range comparisonFields {
			if m.fileMetaData.metaData[i][0] == field {
				return true
			} else if m.fileMetaData.metaData[j][0] == field {
				return false
			}
		}

		// Default comparison
		return m.fileMetaData.metaData[i][0] < m.fileMetaData.metaData[j][0]
	})
	for _, data := range m.fileMetaData.metaData {
		if len(data[0]) > maxKeyLength {
			maxKeyLength = len(data[0])
		}
	}

	sprintfLength := maxKeyLength + 1
	vauleLength := BottomWidth(m.fullWidth) - maxKeyLength - 2
	if vauleLength < BottomWidth(m.fullWidth)/2 {
		vauleLength = BottomWidth(m.fullWidth)/2 - 2
		sprintfLength = vauleLength
	}

	for i := m.fileMetaData.renderIndex; i < bottomElementHight(footerHeight)+m.fileMetaData.renderIndex && i < len(m.fileMetaData.metaData); i++ {
		if i != m.fileMetaData.renderIndex {
			metaDataBar += "\n"
		}
		data := TruncateMiddleText(m.fileMetaData.metaData[i][1], vauleLength)
		metadataName := m.fileMetaData.metaData[i][0]
		if BottomWidth(m.fullWidth)-maxKeyLength-3 < BottomWidth(m.fullWidth)/2 {
			metadataName = TruncateMiddleText(m.fileMetaData.metaData[i][0], vauleLength)
		}
		metaDataBar += fmt.Sprintf("%-*s %s", sprintfLength, metadataName, data)

	}
	bottomBorder := GenerateBottomBorder(fmt.Sprintf("%s/%s", strconv.Itoa(m.fileMetaData.renderIndex+1), strconv.Itoa(len(m.fileMetaData.metaData))), BottomWidth(m.fullWidth)-3)
	metaDataBar = MetaDataBoarder(bottomElementHight(footerHeight), BottomWidth(m.fullWidth), bottomBorder, m.focusPanel).Render(metaDataBar)

	return metaDataBar
}

func ClipboardRender(m model) string {

	// render
	clipboardRender := ""
	if len(m.copyItems.items) == 0 {
		clipboardRender += "\n   No content in clipboard"
	} else {
		for i := 0; i < len(m.copyItems.items) && i < bottomElementHight(footerHeight); i++ {
			if i == bottomElementHight(footerHeight)-1 {
				clipboardRender += strconv.Itoa(len(m.copyItems.items)-i+1) + " item left...."
			} else {
				fileInfo, err := os.Stat(m.copyItems.items[i])
				if err != nil {
					outPutLog("Clipboard render function get item state error", err)
				}
				if !os.IsNotExist(err) {
					clipboardRender += ClipboardPrettierName(m.copyItems.items[i], BottomWidth(m.fullWidth)-3, fileInfo.IsDir(), false) + "\n"
				}
			}
		}
	}
	for i := 0; i < len(m.copyItems.items); i++ {

	}
	bottomWidth := 0

	if m.fullWidth%3 != 0 {
		bottomWidth = BottomWidth(m.fullWidth + m.fullWidth%3 + 2)
	} else {
		bottomWidth = BottomWidth(m.fullWidth)
	}
	clipboardRender = ClipboardBoarder(bottomElementHight(footerHeight), bottomWidth, "━").Render(clipboardRender)

	return clipboardRender
}

func TerminalSizeWarnRender(m model) string {
	focusedModelStyle := lipgloss.NewStyle().
		Height(m.fullHeight).
		Width(m.fullWidth).
		Align(lipgloss.Center, lipgloss.Center).
		BorderForeground(lipgloss.Color("69"))
	fullWidthString := strconv.Itoa(m.fullWidth)
	fullHeightString := strconv.Itoa(m.fullHeight)
	minimumWidthString := strconv.Itoa(minimumWidth)
	minimumHeightString := strconv.Itoa(minimumHeight)
	if m.fullHeight < minimumHeight {
		fullHeightString = terminalTooSmall.Render(fullHeightString)
	}
	if m.fullWidth < minimumWidth {
		fullWidthString = terminalTooSmall.Render(fullWidthString)
	}
	fullHeightString = terminalMinimumSize.Render(fullHeightString)
	fullWidthString = terminalMinimumSize.Render(fullWidthString)

	return focusedModelStyle.Render(`Terminal size too small:` + "\n" +
		"Width = " + fullWidthString +
		" Height = " + fullHeightString + "\n\n" +

		"Needed for current config:" + "\n" +
		"Width = " + terminalMinimumSize.Render(minimumWidthString) +
		" Height = " + terminalMinimumSize.Render(minimumHeightString))
}

func TypineModalRender(m model) string {
	if m.typingModal.itemType == newFile {
		fileLocation := filePanelTopFolderIcon.Render("   ") + filePanelTopPath.Render(TruncateTextBeginning(m.typingModal.location+"/"+m.typingModal.textInput.Value(), modalWidth-4)) + "\n"
		confirm := modalConfirm.Render(" (" + hotkeys.Confirm[0] + ") New File ")
		cancel := modalCancel.Render(" (" + hotkeys.Cancel[0] + ") Cancel ")
		tip := confirm + lipgloss.NewStyle().Background(mainBackgroundColor).Render("           ") + cancel
		return FullScreenStyle(m.fullHeight, m.fullWidth).Render(FocusedModalStyle(modalHeight, modalWidth).Render(fileLocation + "\n" + m.typingModal.textInput.View() + "\n\n" + tip))
	} else {
		fileLocation := filePanelTopFolderIcon.Render("   ") + filePanelTopPath.Render(TruncateTextBeginning(m.typingModal.location+"/"+m.typingModal.textInput.Value(), modalWidth-4)) + "\n"
		confirm := modalConfirm.Render(" (" + hotkeys.Confirm[0] + ") New Folder ")
		cancel := modalCancel.Render(" (" + hotkeys.Cancel[0] + ") Cancel ")
		tip := confirm + lipgloss.NewStyle().Background(mainBackgroundColor).Render("           ") + cancel
		return FullScreenStyle(m.fullHeight, m.fullWidth).Render(FocusedModalStyle(modalHeight, modalWidth).Render(fileLocation + "\n" + m.typingModal.textInput.View() + "\n\n" + tip))
	}
}

func WarnModalRender(m model) string {
	title := m.warnModal.title
	content := m.warnModal.content
	confirm := modalCancel.Render(" (" + hotkeys.Confirm[0] + ") Confirm ")
	cancel := modalCancel.Render(" (" + hotkeys.Cancel[0] + ") Cancel ")
	tip := confirm + lipgloss.NewStyle().Background(mainBackgroundColor).Render("           ") + cancel
	return FullScreenStyle(m.fullHeight, m.fullWidth).Render(FocusedModalStyle(modalHeight, modalWidth).Render(title + "\n\n" + content + "\n\n" + tip))
}
