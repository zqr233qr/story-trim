/**
 * TTS 听书核心逻辑 (纯净版)
 * 仅保留 nrb-tts-plugin 核心功能：
 * 1. 朗读 & 连播 (支持后台)
 * 2. 语速调节
 * 3. 定时关闭
 * 4. 暂停/恢复
 */

export interface TTSOptions {
  onRangeStart?: (index: number) => void;
  onEnd?: () => void;
  onNext?: () => void;
  onPrev?: () => void;
  rate?: number;
  title?: string;
  singer?: string;
  cover?: string;
}

let tts: any = null;

export class TTSPlayer {
  private lines: string[] = [];
  private currentIndex = 0;
  private options: TTSOptions = {};
  
  public isSpeaking = false;
  public isPaused = false;
  
  private sleepTimer: any = null;
  public sleepMode: 'off' | number = 'off';

  constructor(options: TTSOptions = {}) {
    this.options = { rate: 1.0, ...options };
    this.initPlugin();
  }
  
  private initPlugin() {
    // #ifdef APP-PLUS
    if (uni.getSystemInfoSync().platform === 'android') {
      tts = uni.requireNativePlugin("nrb-tts-plugin");
      if (tts) {
        tts.init({ "lang": "ZH", "country": "CN" }, (res: any) => {});
      }
    }
    // #endif
  }

  setLines(lines: string[], startIndex = 0) {
    this.lines = lines;
    this.currentIndex = startIndex;
  }
  
  updateOptions(options: Partial<TTSOptions>) {
    this.options = { ...this.options, ...options };
    if (this.isSpeaking) {
       this.play(this.currentIndex); 
    }
  }
  
  setSleepTimer(mode: 'off' | number) {
    this.sleepMode = mode;
    clearTimeout(this.sleepTimer);
    if (typeof mode === 'number' && mode > 0) {
      console.log(`TTS: Sleep timer ${mode} min`);
      this.sleepTimer = setTimeout(() => { this.stop(); }, mode * 60 * 1000);
    }
  }

  play(index?: number) {
    if (index !== undefined) this.currentIndex = index;
    this.isSpeaking = true;
    this.isPaused = false;

    // #ifdef APP-PLUS
    if (tts) this.playStep();
    // #endif
  }
  
  private playStep() {
    if (this.currentIndex >= this.lines.length) {
      this.isSpeaking = false;
      this.options.onEnd?.(); 
      return;
    }

    const text = this.lines[this.currentIndex];
    if (!text || text.trim().length === 0) {
      this.currentIndex++;
      this.playStep();
      return;
    }

    const rate = this.options.rate || 1.0;
    const params = {
      "speechRate": rate,
      "rate": rate,
      "pitch": 1.0,
      "queueMode": 1
    };

    tts.speak(text, params, (e: any) => {
      if (e.type === 'onStart') {
        this.options.onRangeStart?.(this.currentIndex);
      } 
      else if (e.type === 'onDone') {
        if (this.isSpeaking && !this.isPaused) {
          this.currentIndex++;
          this.playStep();
        }
      } 
      else if (e.type === 'onError') {
        // 遇到错误跳过当前句，继续下一句
        if (this.isSpeaking) {
            this.currentIndex++;
            this.playStep();
        }
      }
    });
  }

  pause() {
    this.isPaused = true;
    this.isSpeaking = false;
    if (tts) tts.stop();
  }

  resume() {
    this.play();
  }

  stop() {
    this.isSpeaking = false;
    this.isPaused = false;
    clearTimeout(this.sleepTimer);
    this.sleepMode = 'off';
    if (tts) tts.stop();
  }
}