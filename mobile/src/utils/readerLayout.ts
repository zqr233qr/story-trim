/* eslint-disable */
const inBrowser = typeof window !== 'undefined'
const baseChar = '阅'
let lineH: Record<string, number> = {}
let remainHeight = 0
let options = {
  platform: 'browser',
  id: '',
  splitCode: '\n',
  fast: false,
  type: 'page',
  width: 0,
  height: 0,
  fontFamily: 'sans-serif',
  fontSize: 0,
  lineHeight: 1.4,
  pGap: 0,
  title: '',
  titleSize: 0,
  titleHeight: 1.4,
  titleWeight: 'normal',
  titleGap: 0
}

let cacheData = {
  cWidth: 0,
  cHeight: 0,
  cfontSize: 0,
  maxText: 0,
  maxLine: 0
}

type Options = typeof options

type PageListItem = {
  isTitle: boolean
  center: boolean
  pFirst: boolean
  pIndex: number
  lineIndex: number
  textIndex: number
  text: string
}

type PageList = PageListItem[]

const getStyle = (attr: string) => {
  if (!inBrowser) {
    return ''
  }
  if (getComputedStyle) {
    const styles: any = getComputedStyle(document.documentElement)
    return styles[attr] || ('Arial' as string)
  }
  return 'Arial'
}

const trimAll = (str: string) => {
  if (str) {
    return String(str).replace(/[\s]+/gim, '')
  }
  return ''
}

export function Reader(content: string, option: Options) {
  const { type, width, height, fontFamily, fontSize, title, titleSize } = option
  if (!content) {
    console.warn('无内容')
    return []
  }
  if (!width || Number(width) <= 0) {
    console.warn('请传入容器宽度，值需要大于 0')
    return []
  }
  if (type === 'page' && (!height || Number(height) <= 0)) {
    console.warn('请传入容器高度，值需要大于 0')
    return []
  }
  if (!fontSize || Number(fontSize) <= 0) {
    console.warn('请传入章节内容字号大小，值需要大于 0')
    return []
  }
  if (title && (!titleSize || Number(titleSize) <= 0)) {
    console.warn('请传入章节标题字号大小，值需要大于 0')
    return []
  }
  options = { ...options, ...option }

  lineH = {}

  const { cWidth, cHeight, cfontSize } = cacheData
  if (cWidth !== width || cHeight !== height || cfontSize !== fontSize) {
    cacheData = {
      cWidth: width,
      cHeight: height,
      cfontSize: fontSize,
      maxText: 0,
      maxLine: 0
    }
  }

  const rootFamily = getStyle('font-family')
  if (!fontFamily && rootFamily) {
    options.fontFamily = rootFamily
  }

  if (type === 'line') {
    return splitContent2lines(content)
  }

  const lines = splitContent2lines(content)
  return joinLine2Pages(lines)
}

function splitContent2lines(content: string) {
  const { splitCode, width, fontSize, title } = options

  let hasTitle = false
  const reg = `[${splitCode}]+`
  const pList = content
    .split(new RegExp(reg, 'gim'))
    .map((v, i) => {
      if (i === 0 && v === title) {
        hasTitle = true
        return v
      }
      return trimAll(v)
    })
    .filter((v) => v)

  if (!hasTitle) {
    pList.unshift(title)
  }
  if (title && trimAll(pList[1]) === trimAll(title)) {
    pList.splice(1, 1)
  }

  if (!cacheData.maxText) {
    const baseLen = Math.floor(width / fontSize)
    let char = ''
    for (let i = 0; i < baseLen; i++) {
      char += baseChar
    }
    const maxText = getText({ fontSize }, char, true)
    cacheData.maxText = maxText.length
  }

  let result: PageList = []
  pList.forEach((pText, index) => {
    result = result.concat(p2line(pText, index, cacheData.maxText))
  })

  return result
}

function p2line(pText: string, index: number, maxLen: number) {
  const { fast, fontSize, title, titleSize, titleWeight } = options
  const isTitle = pText === title
  let p = pText
  let tag = 0
  let lines: PageList = []

  while (p) {
    tag += 1
    const pFirst = !isTitle && tag === 1
    const sliceLen = pFirst ? maxLen - 2 : maxLen
    let lineText = p.slice(0, sliceLen)
    if (pFirst) {
      lineText = baseChar + baseChar + lineText
    }

    if (!isTitle && p.length <= sliceLen) {
      p = ''
    } else {
      if (!fast || isTitle) {
        lineText = getText(
          {
            p,
            sliceLen,
            fontSize: isTitle ? titleSize : fontSize,
            weight: isTitle ? titleWeight : ''
          },
          lineText
        )
      }
      p = p.slice(pFirst ? lineText.length - 2 : lineText.length)
    }

    if (pFirst) {
      lineText = lineText.slice(2)
    }

    let center = true
    if (p) {
      const { transLine, transP, canCenter } = transDot(lineText, p)
      lineText = transLine
      p = transP
      center = canCenter
    }
    if (p) {
      const { transLine, transP, canCenter } = transNumEn(lineText, p, center)
      lineText = transLine
      p = transP
      center = canCenter
    }

    if (isTitle || !p) {
      center = false
    }

    lines.push({
      isTitle,
      center,
      pFirst: !isTitle && tag === 1,
      pIndex: index,
      lineIndex: tag,
      textIndex: pText.indexOf(lineText),
      text: lineText
    })
  }
  return lines
}

function getText(
  params: { p?: string; sliceLen?: number; fontSize: number; weight?: number | string },
  text: string,
  base = false,
  fontW?: number
) {
  const { width, fontFamily } = options
  const { p, sliceLen, fontSize, weight } = params
  const getWidth = (text: string) => {
    return getTextWidth(text, fontSize, fontFamily, weight)
  }

  const textW = fontW || getWidth(text)
  if (textW === width) {
    return text
  }

  if (textW < width) {
    const add = p && p.slice(sliceLen, sliceLen && sliceLen + 1)
    if (!base && !add) {
      return text
    }
    const addText = base ? text + baseChar : text + add
    const addTextW = getWidth(addText)
    if (addTextW === width) {
      return addText
    }
    if (addTextW > width) {
      return text
    }
    return getText({ ...params, sliceLen: sliceLen && sliceLen + 1 }, addText, base, addTextW)
  }

  const cutText = text.slice(0, -1)
  if (!cutText) {
    return text
  }
  const cutTextW = getWidth(cutText)
  if (cutTextW <= width) {
    return cutText
  }
  return getText(params, cutText, base, cutTextW)
}

let canvas: HTMLCanvasElement | null = null
let ctx: CanvasRenderingContext2D | null = null
function getTextWidth(text: string, fontSize: number, fontFamily: string, weight?: number | string) {
  if (!canvas) {
    canvas = document.createElement('canvas')
    ctx = canvas.getContext('2d')
  }

  ctx!.font = `${weight ? weight : 'normal'} ${fontSize}px PingFang SC`
  const { width } = ctx!.measureText(text)
  return width
}

function joinLine2Pages(lines: PageList) {
  const { height } = options
  if (!cacheData.maxLine) {
    let maxLine = 1
    if (lines.length >= 2) {
      const baseLineH = getLineHeight(lines[1], 0, 'base')
      remainHeight = (1 / 3) * baseLineH
      maxLine = Math.floor(height / baseLineH)
    }
    cacheData.maxLine = maxLine
  }

  let pageLines = lines.slice(0)
  let pages: PageList[] = []
  while (pageLines.length > 0) {
    const page = getPage(pageLines, cacheData.maxLine)
    pages.push(page)
    pageLines = pageLines.slice(page.length)
  }

  return pages
}

function getPage(lines: PageList, maxLine: number, pageHeight?: number) {
  const { height, titleGap } = options
  const page = lines.slice(0, maxLine)
  const pageH = pageHeight || getPageHeight(page)
  let contHeight = height
  if (lines && lines[0] && lines[0].isTitle) {
    contHeight = height - titleGap
  }

  if (pageH === contHeight) {
    return page
  }

  if (pageH < contHeight + remainHeight) {
    const add = maxLine + 1
    const addLine = lines.slice(maxLine, add)
    if (addLine.length <= 0) {
      return page
    }
    const addPage = lines.slice(0, add)
    const addPageH = getPageHeight(addPage)
    if (addPageH === contHeight) {
      return addPage
    }
    if (addPageH > contHeight) {
      freedLineH(addLine[0])
      return page
    }
    return getPage(lines, add, addPageH)
  }

  const cut = maxLine - 1
  if (cut <= 0) {
    return page
  }
  const cutPage = lines.slice(0, cut)
  const cutPageH = getPageHeight(cutPage)
  if (cutPageH <= contHeight) {
    freedLineH(lines.slice(cut, maxLine)[0])
    return cutPage
  }
  return getPage(lines, cut, cutPageH)
}

function freedLineH(line: PageListItem) {
  const tempKey = `${line.pIndex}_${line.lineIndex}`
  lineH[tempKey] = 0
}

function getLineHeight(line: PageListItem, linesIndex: number, type?: 'base') {
  const index = `${line.pIndex}_${line.lineIndex}`
  let theLineH = lineH[index]
  if (theLineH) {
    return theLineH
  }

  const { pGap, fontSize, lineHeight, titleSize, titleHeight } = options
  const size = line.isTitle ? titleSize : fontSize
  const height = line.isTitle ? titleHeight : lineHeight

  if (type === 'base') {
    return fontSize * lineHeight
  }

  let gap = 0
  if (!line.isTitle && line.lineIndex === 1 && linesIndex !== 0) {
    gap = pGap
  }
  theLineH = size * height + gap
  lineH[index] = theLineH
  return theLineH
}

function getPageHeight(lines: PageList) {
  let pageH = 0
  lines.forEach((line, index: number) => {
    pageH += getLineHeight(line, index)
  })
  return pageH
}

function transDot(line: string, p: string) {
  let transLine = line
  let transP = p
  let canCenter = true

  if (isDot(p.slice(0, 1))) {
    transLine = line.slice(0, -1)
    transP = line.slice(-1) + p

    const endDot = getEndDot(line)
    if (endDot && endDot.length > 0) {
      let len = endDot.length
      if (len >= 3 || len >= line.length - 2) {
        return { transLine: line, transP: p, canCenter: true }
      }
      len = len + 1
      transLine = line.slice(0, -len)
      transP = line.slice(-len) + p
      canCenter = false
    }
  }

  return { transLine, transP, canCenter }
}

function transNumEn(line: string, p: string, center: boolean) {
  const pFirst = p.slice(0, 1)
  let transLen = 0
  let transLine = line
  let transP = p
  let canCenter = center

  if (/\d/gi.test(pFirst)) {
    const endNum = getEndNum(line)
    if (endNum && endNum.length > 0) {
      const len = endNum[0].length
      if (len < line.length) {
        transLen = len
      }
    }
  } else if (/[a-zA-Z]/gi.test(pFirst)) {
    const endEn = getEndEn(line)
    if (endEn && endEn.length > 0) {
      const len = endEn[0].length
      if (len < line.length) {
        transLen = len
      }
    }
  }
  if (transLen) {
    transLine = line.slice(0, -transLen)
    transP = line.slice(-transLen) + p
    canCenter = false
  }

  return { transLine, transP, canCenter }
}

function isDot(code: string) {
  if (!code) {
    return false
  }
  const dots = [
    'ff0c',
    '3002',
    'ff1a',
    'ff1b',
    'ff01',
    'ff1f',
    '3001',
    'ff09',
    '300b',
    '300d',
    '3011',
    '2c',
    '2e',
    '3a',
    '3b',
    '21',
    '3f',
    '5e',
    '29',
    '3e',
    '7d',
    '5d',
    '2026',
    '7e',
    '25',
    'b7',
    '2019',
    '201d',
    '60',
    '2d',
    '2014',
    '5f',
    '7c',
    '5c',
    '2f'
  ]
  const charCode = code.charCodeAt(0).toString(16)
  return dots.includes(charCode)
}

function getEndDot(str: string) {
  return str.match(
    /[\uff0c|\u3002|\uff1a|\uff1b|\uff01|\uff1f|\u3001|\uff09|\u300b|\u300d|\u3011|\u002c|\u002e|\u003a|\u003b|\u0021|\u003f|\u005e|\u0029|\u003e|\u007d|\u005d|\u2026|\u007e|\u0025|\u00b7|\u2019|\u201d|\u0060|\u002d|\u2014|\u005f|\u007c|\u005c|\u002f\uff08|\u300a|\u300c|\u3010|\u0028|\u003c|\u007b|\u005b|\u2018|\u201c|\u0040|\u0023|\uffe5|\u0024|\u0026]+$/gi
  )
}

function getEndNum(str: string) {
  return str.match(/[0-9]+$/gi)
}

function getEndEn(str: string) {
  return str.match(/[a-zA-Z]+$/gi)
}
