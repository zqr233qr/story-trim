import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface Chapter {
  id: number;
  index: number;
  title: string;
  modes: Record<string, string[]>; // modeId -> content lines
}

export interface Book {
  id: number;
  title: string;
  cover?: string;
  totalChapters: number;
  progress: number; // 0-100
  status: 'new' | 'processing' | 'ready';
  activeModeId?: string; // 当前激活的 AI 模式 ID
  
  // Runtime State (UI use)
  activeChapterIndex: number;
  chapters: Chapter[];
}

// Helper to generate long text
const generateLongText = (base: string[]) => {
  let res = [...base]
  for (let i = 0; i < 5; i++) { // Repeat 5 times to ensure scroll
    res = res.concat(base)
  }
  return res
}

export const useBookStore = defineStore('book', () => {
  
  // 预置三章内容
  const mockChapters: Chapter[] = [
    {
      id: 101, index: 0, title: '第一章 陨落的天才',
      modes: {
        original: generateLongText([
          "“斗之力，三段！”",
          "望着测验魔石碑上面闪亮得甚至有些刺眼的五个大字，少年面无表情，唇角有着一抹自嘲，紧握的手掌，因为大力，而导致略微尖锐的指甲深深的刺进了掌心之中，带来一阵阵钻心的疼痛…",
          "“萧炎，斗之力，三段！级别：低级！”测验魔石碑之旁，一位中年男子，看了一眼碑上所显示出来的信息，语气漠然的将之公布了出来…",
          "中年男子话刚刚脱口，便是不出意外的在人头汹涌的广场上带起了一阵嘲讽的骚动。",
          "“三段？嘿嘿，果然不出我所料，这个‘天才’这一年又是在原地踏步！”",
          "“哎，这废物真是把家族的脸都给丢光了。”",
          "“要不是族长是他的父亲，这种废物，早就被驱赶出家族，任其自生自灭了，哪还有机会待在家族中白吃白喝。”",
          "“唉，昔年那名闻乌坦城的天才少年，如今怎么落魄成这般模样了？”"
        ]),
        dewater: generateLongText([
          "“斗之力，三段！”",
          "测验魔石碑上闪烁着刺眼的五个大字。少年面无表情，唇角勾起一抹自嘲，指甲深深嵌入掌心，带来钻心的痛。",
          "“萧炎，斗之力，三段！级别：低级！”测验员漠然公布。",
          "广场上顿时响起一阵嘲讽的骚动。",
          "“三段？果然，这个‘天才’今年还在原地踏步！”",
          "“要不是族长是他父亲，这种废物早就被赶出去了。”"
        ]),
        summary: ["萧炎测试结果为斗之力三段，遭受族人嘲讽。"]
      }
    },
    {
      id: 102, index: 1, title: '第二章 斗气大陆',
      modes: {
        original: generateLongText([
          "月如银盘，漫天繁星。",
          "山崖之顶，萧炎斜躺在草地之上，嘴中叼中一根青草，微微嚼动，任由那淡淡的苦涩在嘴中弥漫开来…",
          "举起白皙的手掌，挡在眼前，目光透过手指缝隙，遥望着天空上那轮巨大的银月。",
          "“唉…”叹了一口气，懒懒的抽回手掌，双手枕着脑后，眼神有些恍惚…",
          "“十五年了呢…”低低的自喃声，忽然毫无边际的从少年嘴中轻吐了出来。",
          "在萧炎的心中，有一个仅有他自己知道的秘密：他并不是这个世界的人，或者说，萧炎的灵魂，并不属于这个世界，他来自一个名叫地球的蔚蓝星球，至于为什么会来到这里，这种离奇经过，他也无法解释，不过在生活了一段时间之后，他还是慢慢的了解到：这片大陆，名叫斗气大陆…"
        ]),
        dewater: generateLongText([
          "月如银盘。山崖之顶，萧炎叼着青草，仰望银月。",
          "“十五年了…”",
          "萧炎有一个秘密：他穿越自地球。这片大陆名为斗气大陆，没有魔法，只有繁衍到巅峰的斗气。"
        ]),
        summary: ["萧炎独自在后山感慨穿越十五年的经历，介绍了斗气大陆的世界观。"]
      }
    },
    {
      id: 103, index: 2, title: '第三章 纳兰嫣然',
      modes: {
        original: generateLongText([
          "大厅之中，气氛略微有些沉闷。",
          "三位长老面面相觑，最后目光都是投向了首位之上那位脸色有些难看的中年人身上，无可奈何的摇了摇头。",
          "“族长，那纳兰侄女…喔，不，纳兰小姐，还是先把婚事退了吧。”三长老有些阴测测的道。",
          "“老三，你闭嘴！”萧战一拍桌子，怒喝道。",
          "“嘿嘿，族长，你也别对我发火，这可是云岚宗的意思…”"
        ]),
        dewater: generateLongText([
          "大厅气氛沉闷。",
          "三位长老看向脸色难看的族长萧战。",
          "三长老阴测测道：“族长，先把纳兰小姐的婚退了吧。”",
          "萧战怒喝：“闭嘴！”",
          "“这可是云岚宗的意思。”"
        ]),
        summary: ["纳兰嫣然上门退婚，萧战与长老发生冲突。"]
      }
    }
  ]

  const books = ref<Book[]>([
    {
      id: 1,
      title: '斗破苍穹',
      totalChapters: 3,
      progress: 0,
      status: 'ready',
      activeModeId: 'dewater',
      activeChapterIndex: 0,
      chapters: mockChapters
    }
  ])

  const activeBook = ref<Book | null>(null)

  const setActiveBook = (bookId: number) => {
    activeBook.value = books.value.find(b => b.id === bookId) || null
  }

  // 切换章节
  const setChapter = (index: number) => {
    if (activeBook.value && index >= 0 && index < activeBook.value.chapters.length) {
      activeBook.value.activeChapterIndex = index
      activeBook.value.progress = Math.floor((index / activeBook.value.chapters.length) * 100)
    }
  }

  const updateBookStatus = (bookId: number, status: 'new' | 'processing' | 'ready') => {
    const book = books.value.find(b => b.id === bookId)
    if (book) book.status = status
  }

  const activeChapter = computed(() => {
    if (!activeBook.value) return null
    return activeBook.value.chapters[activeBook.value.activeChapterIndex]
  })

  return { books, activeBook, activeChapter, setActiveBook, setChapter, updateBookStatus }
})