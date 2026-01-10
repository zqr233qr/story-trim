import { BASE_URL } from './index';

/**
 * 基于章节ID的流式精简
 * 对应后端: GET /trim/stream/by-id (WebSocket)
 * 
 * 适用场景: 小程序、App降级模式
 * 
 * @param chapterId 章节ID
 * @param promptId 精简模式ID
 * @param onData 接收文本片段的回调
 * @param onError 错误回调
 * @param onDone 完成回调
 */
export function trimStreamByChapterId(
  chapterId: number,
  promptId: number,
  onData: (text: string) => void,
  onError: (err: string) => void,
  onDone: () => void
) {
  const token = uni.getStorageSync('token');
  const wsUrl = `${BASE_URL.replace('http', 'ws')}/trim/stream/by-id`;
  
  console.log('[TrimStream] Connecting to:', wsUrl);
  
  const socketTask = uni.connectSocket({ url: wsUrl });
  
  socketTask.onOpen(() => {
    console.log('[TrimStream] Connected, sending request...');
    // 发送认证信息
    socketTask.send({
      data: JSON.stringify({
        token,
        chapter_id: chapterId,
        prompt_id: promptId
      })
    });
  });
  
  socketTask.onMessage((res) => {
    try {
      const data = JSON.parse(res.data as string);
      if (data.error) {
        console.error('[TrimStream] Server error:', data.error);
        onError(data.error);
        socketTask.close({});
      } else if (data.c) {
        onData(data.c);
      }
    } catch (e) {
      console.error('[TrimStream] Parse error:', e);
    }
  });
  
  socketTask.onClose(() => {
    console.log('[TrimStream] Connection closed');
    onDone();
  });
  
  socketTask.onError((err) => {
    console.error('[TrimStream] Connection error:', err);
    onError('WebSocket 连接失败');
  });
  
  return socketTask;
}

/**
 * 基于MD5的流式精简
 * 对应后端: GET /trim/stream/by-md5 (WebSocket)
 * 
 * 适用场景: App已同步模式（离线优先）
 * 
 * @param content 章节原文
 * @param md5 章节MD5
 * @param promptId 精简模式ID
 * @param bookFingerprint 书籍指纹
 * @param chapterIndex 章节索引
 * @param onData 接收文本片段的回调
 * @param onError 错误回调
 * @param onDone 完成回调
 */
export function trimStreamByMd5(
  content: string,
  md5: string,
  promptId: number,
  bookFingerprint: string,
  chapterIndex: number,
  onData: (text: string) => void,
  onError: (err: string) => void,
  onDone: () => void
) {
  const token = uni.getStorageSync('token');
  const wsUrl = `${BASE_URL.replace('http', 'ws')}/trim/stream/by-md5`;
  
  console.log('[TrimStream] Connecting to:', wsUrl);
  
  const socketTask = uni.connectSocket({ url: wsUrl });
  
  socketTask.onOpen(() => {
    console.log('[TrimStream] Connected, sending request...');
    socketTask.send({
      data: JSON.stringify({
        token,
        content,
        md5,
        prompt_id: promptId,
        book_fingerprint: bookFingerprint,
        chapter_index: chapterIndex
      })
    });
  });
  
  socketTask.onMessage((res) => {
    try {
      const data = JSON.parse(res.data as string);
      if (data.error) {
        console.error('[TrimStream] Server error:', data.error);
        onError(data.error);
        socketTask.close({});
      } else if (data.c) {
        onData(data.c);
      }
    } catch (e) {
      console.error('[TrimStream] Parse error:', e);
    }
  });
  
  socketTask.onClose(() => {
    console.log('[TrimStream] Connection closed');
    onDone();
  });
  
  socketTask.onError((err) => {
    console.error('[TrimStream] Connection error:', err);
    onError('WebSocket 连接失败');
  });
  
  return socketTask;
}
