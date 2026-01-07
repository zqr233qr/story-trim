const switchToMode = async (id: string, showModalOnFailure = true) => {
  if (activeChapter.value?.modes[id]) {
    activeBook.value!.activeModeId = id
    isMagicActive.value = true
    if (pageMode.value === 'click') refreshWindow('keep')
    return
  }
  isAiLoading.value = true
  const success = await bookStore.fetchChapterTrim(activeChapter.value!.id, parseInt(id))
  isAiLoading.value = false
  if (success) {
    activeBook.value!.activeModeId = id
    isMagicActive.value = true
    if (pageMode.value === 'click') refreshWindow('keep')
  } else if (showModalOnFailure) {
    showConfigModal.value = true
  }
}

const toggleMagic = () => {
  if (isMagicActive.value) {
    isMagicActive.value = false
    if (pageMode.value === 'click') refreshWindow('keep')
  } else {
    const targetMode = activeBook.value?.activeModeId || (bookStore.prompts[0]?.id.toString())
    if (targetMode) {
      switchToMode(targetMode, true)
    } else {
      showConfigModal.value = true
    }
  }
}

const onSwiperChange = (e: any) => {
  if (!isSwiperReady.value || isWindowShifting.value) return
  const newIdx = e.detail.current
  const prevCount = prevPages.value.length
  const currCount = currPages.value.length
  
  if (newIdx < prevCount) {
    isWindowShifting.value = true
    bookStore.activeBook!.activeChapterIndex -= 1
    refreshWindow('last').then(() => { isWindowShifting.value = false })
  } else if (newIdx >= prevCount + currCount) {
    isWindowShifting.value = true
    bookStore.activeBook!.activeChapterIndex += 1
    refreshWindow('first').then(() => { isWindowShifting.value = false })
  } else {
    swiperCurrent.value = newIdx
    currentPageIndex.value = newIdx
  }
}

const handleContentClick = (e: any) => {
  if (menuVisible.value) { menuVisible.value = false; return }
  if (pageMode.value === 'click') {
    const info = uni.getSystemInfoSync()
    const x = e.detail.x
    if (x < info.windowWidth * 0.3) {
      if (swiperCurrent.value > 0) swiperCurrent.value--
      else if (activeBook.value!.activeChapterIndex > 0) switchChapter(activeBook.value!.activeChapterIndex - 1, 'end')
    } else if (x > info.windowWidth * 0.7) {
      if (swiperCurrent.value < combinedPages.value.length - 1) swiperCurrent.value++
      else switchChapter(activeBook.value!.activeChapterIndex + 1, 'start')
    } else { menuVisible.value = true }
  } else { menuVisible.value = true }
}

const switchChapter = async (index: number, targetPosition: 'start' | 'end' = 'start') => {
  if (index < 0 || index >= activeBook.value!.chapters.length) return
  isTextTransitioning.value = true
  if (pageMode.value === 'scroll') {
    scrollTop.value = 1
    nextTick(() => { scrollTop.value = 0 })
  }
  await bookStore.setChapter(index)
  if (isMagicActive.value && activeBook.value && activeChapter.value) {
    const modeId = activeBook.value.activeModeId
    if (modeId && !activeChapter.value.modes[modeId]) {
      const isTrimmed = activeChapter.value.trimmed_prompt_ids?.includes(Number(modeId))
      if (isTrimmed) await bookStore.fetchChapterTrim(activeChapter.value.id, Number(modeId))
    }
    if (modeId && !activeChapter.value.modes[modeId]) {
      isMagicActive.value = false
      const p = bookStore.prompts.find(p => p.id.toString() === modeId)
      showNotification(`本章暂无[${p?.name || 'AI'}]数据，已切回原文`)
    }
  }
  if (pageMode.value === 'click') { rePaginate(targetPosition) } 
  else { isTextTransitioning.value = false }
}

const rePaginate = async (targetPosition: 'start' | 'end' = 'start') => {
  if (pageMode.value !== 'click') return
  isTextTransitioning.value = true
  isSwiperReady.value = false
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 150))
  const info = uni.getSystemInfoSync()
  const viewHeight = info.windowHeight - 160
  const query = uni.createSelectorQuery().in(instance)
  query.selectAll('.measurer-para').boundingClientRect()
  query.exec(async (res) => {
    if (!res || !res[0]) { isTextTransitioning.value = false; return }
    const rects = res[0] as any[]
    let currentPage: string[] = []
    let currentHeight = 0
    const pages: string[][] = []
    if (rects.length > 0) {
      rects.forEach((rect, idx) => {
        const paraText = currentText.value[idx]
        const h = rect ? rect.height : 40
        if (currentHeight + h > viewHeight && currentPage.length > 0) {
          pages.push(currentPage)
          currentPage = [paraText]
          currentHeight = h
        } else {
          currentPage.push(paraText)
          currentHeight += h + 24
        }
      })
      if (currentPage.length > 0) pages.push(currentPage)
    }
    const newPages = pages.length > 0 ? pages : [currentText.value]
    currPages.value = newPages 
    const targetIdx = targetPosition === 'end' ? newPages.length - 1 : 0
    swiperCurrent.value = targetIdx
    currentPageIndex.value = targetIdx
    await nextTick()
    isSwiperReady.value = true
    isTextTransitioning.value = false
  })
}
</script>