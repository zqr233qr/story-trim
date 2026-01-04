import axios from 'axios';

const BASE_URL = 'http://localhost:8080/api';

const client = axios.create({
  baseURL: BASE_URL,
  timeout: 300000, 
});

export interface Chapter {
  index: number;
  title: string;
  content: string;
}

export interface UploadResponse {
  filename: string;
  chapters: Chapter[];
  total: number;
}

export const api = {
  upload: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return client.post<any>('/upload', formData);
  },
  
  trim: (content: string) => {
    return client.post<any>('/trim', { content });
  },

  trimStream: async (
    content: string, 
    onData: (text: string) => void,
    onError: (err: string) => void,
    onDone: () => void
  ) => {
    try {
      const response = await fetch(`${BASE_URL}/trim/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ content }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const reader = response.body?.getReader();
      if (!reader) throw new Error('Response body is null');

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        
        // 简单状态机解析 SSE
        const lines = buffer.split('\n');
        // 如果最后一行不为空，说明 buffer 还没结束，保留到下一轮
        // 但是 SSE 消息以 \n\n 结尾，所以 split 后最后一行通常是空字符串（如果完整）
        // 如果不完整，最后一行就是残留数据
        
        buffer = lines.pop() ?? ''; 

        let currentEvent = '';
        
        for (const line of lines) {
          if (line.startsWith('event:')) {
            currentEvent = line.substring(6).trim();
          } else if (line.startsWith('data:')) {
            const data = line.substring(5); 
            // 如果是 message 事件
            if (currentEvent === 'message' || currentEvent === '') {
               // data: 后面通常有个空格，但如果内容本身是空格开头呢？
               // Gin 实现是 fmt.Fprintf(w, "data: %v\n\n", data)
               // 所以 data 是原样输出的。但通常前面没有空格，除非我们自己加了？
               // 实际上 Gin 的 c.SSEvent 源码：
               // "," err := fmt.Fprintf(c.Writer, "data:%s\n\n", string(data))
               // 不对，是 data:%s (没有空格) —— 抱歉，Gin 1.7+ 改了
               // 让我们假设 data: 后面直接就是内容。但为了安全，如果是以空格开头，只去掉第一个空格。
               
               let text = data;
               if (text.startsWith(' ')) {
                 text = text.substring(1);
               }
               onData(text);
            } else if (currentEvent === 'error') {
               onError(data.trim());
            }
          } else if (line.trim() === '') {
            // 空行意味着事件结束，重置
            currentEvent = '';
          }
        }
      }
      onDone();
    } catch (err: any) {
      onError(err.message || 'Network Error');
    }
  }
};
