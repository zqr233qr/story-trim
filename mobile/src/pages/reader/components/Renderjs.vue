<template>
  <view :prop="rjsChapter" :change:prop="renderScript.loadChapter" style="position: absolute; display: none; top: -9999px; left: -9999px"></view>
</template>

<script lang="ts">
export default {
  props: {
    rjsChapter: {
      type: Object,
      default: () => ({})
    }
  },
  methods: {
    updateContainerContent(params: any) {
      this.$emit('handelContentUpdated', params)
    }
  }
}
</script>

<script module="renderScript" lang="renderjs">
import { Reader } from '@/utils/readerLayout'

export default {
  data() {
    return {
      pageWidth: 0,
      pageHeight: 0,
      statusBarHeight: 0,
      bookOption: {},
      canvas: null
    }
  },
  methods: {
    calcByReader(data) {
      const content = {
        tit: data.title,
        cont: data.content
      }
      const width = this.pageWidth - this.bookOption.sizeInfo.lrPadding * 2
      const height =
        this.pageHeight -
        this.bookOption.sizeInfo.infoHeight -
        this.bookOption.sizeInfo.infoHeight -
        this.bookOption.sizeInfo.tPadding -
        this.bookOption.sizeInfo.bPadding -
        this.statusBarHeight

      const list = Reader(content.cont, {
        platform: 'browser',
        id: 'canvas',
        splitCode: '\n',
        width,
        height,
        fontSize: this.bookOption.sizeInfo.p,
        lineHeight: this.bookOption.sizeInfo.lineHeight,
        pGap: this.bookOption.sizeInfo.margin,
        title: content.tit,
        titleSize: this.bookOption.sizeInfo.title,
        titleHeight: this.bookOption.sizeInfo.titleLineHeight,
        titleWeight: 500,
        titleGap: this.bookOption.sizeInfo.margin,
        fast: true,
        type: this.bookOption.sizeInfo.type
      })

      return list
    },
    loadChapter(newVal) {
      if (newVal?.data?.content) {
        const { pageWidth, pageHeight, statusBarHeight, bookOption } = newVal.options
        this.pageWidth = pageWidth
        this.pageHeight = pageHeight
        this.statusBarHeight = statusBarHeight
        this.bookOption = bookOption
        const d1 = +new Date()
        const chapterPageList = this.calcByReader(newVal.data)
        const d2 = +new Date()
        console.log(d2 - d1, '[Reader][Renderjs] paginate time', {
          type: this.bookOption?.sizeInfo?.type,
          width: pageWidth,
          height: pageHeight,
          contentLength: newVal?.data?.content?.length,
          pages: Array.isArray(chapterPageList) ? chapterPageList.length : -1
        })
        this.$ownerInstance.callMethod('updateContainerContent', { params: newVal, chapterPageList })
      }
    }
  }
}
</script>

<style scoped>
</style>
