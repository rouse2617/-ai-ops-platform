// 增强版 MCP 服务器
// 运行: node server.js

const http = require('http');
const os = require('os');
const dns = require('dns');
const { promisify } = require('util');
const crypto = require('crypto');

const dnsResolve = promisify(dns.resolve);
const PORT = 3001;

// 定义工具
const tools = [
  {
    name: "get_weather",
    description: "获取指定城市的天气信息（模拟数据）",
    inputSchema: {
      type: "object",
      properties: {
        city: { type: "string", description: "城市名称" }
      },
      required: ["city"]
    }
  },
  {
    name: "calculate",
    description: "执行数学计算",
    inputSchema: {
      type: "object",
      properties: {
        expression: { type: "string", description: "数学表达式，如 2+3*4" }
      },
      required: ["expression"]
    }
  },
  {
    name: "get_time",
    description: "获取当前时间，支持多种时区",
    inputSchema: {
      type: "object",
      properties: {
        timezone: { type: "string", description: "时区，如 Asia/Shanghai, America/New_York" }
      }
    }
  },
  {
    name: "system_info",
    description: "获取系统信息（CPU、内存、平台等）",
    inputSchema: {
      type: "object",
      properties: {
        type: {
          type: "string",
          description: "信息类型: all, cpu, memory, platform, uptime",
          enum: ["all", "cpu", "memory", "platform", "uptime"]
        }
      }
    }
  },
  {
    name: "dns_lookup",
    description: "DNS 查询，解析域名获取 IP 地址",
    inputSchema: {
      type: "object",
      properties: {
        domain: { type: "string", description: "要查询的域名" },
        type: { type: "string", description: "记录类型: A, AAAA, MX, TXT, NS", enum: ["A", "AAAA", "MX", "TXT", "NS"] }
      },
      required: ["domain"]
    }
  },
  {
    name: "base64",
    description: "Base64 编码或解码",
    inputSchema: {
      type: "object",
      properties: {
        action: { type: "string", description: "操作类型", enum: ["encode", "decode"] },
        text: { type: "string", description: "要处理的文本" }
      },
      required: ["action", "text"]
    }
  },
  {
    name: "json_format",
    description: "格式化或压缩 JSON",
    inputSchema: {
      type: "object",
      properties: {
        json: { type: "string", description: "JSON 字符串" },
        action: { type: "string", description: "操作类型", enum: ["format", "minify"] }
      },
      required: ["json"]
    }
  },
  {
    name: "uuid_generate",
    description: "生成 UUID",
    inputSchema: {
      type: "object",
      properties: {
        count: { type: "number", description: "生成数量，默认 1" }
      }
    }
  },
  {
    name: "hash",
    description: "计算文本的哈希值",
    inputSchema: {
      type: "object",
      properties: {
        text: { type: "string", description: "要计算哈希的文本" },
        algorithm: { type: "string", description: "哈希算法", enum: ["md5", "sha1", "sha256", "sha512"] }
      },
      required: ["text"]
    }
  },
  {
    name: "random",
    description: "生成随机数或随机字符串",
    inputSchema: {
      type: "object",
      properties: {
        type: { type: "string", description: "类型: number, string, password", enum: ["number", "string", "password"] },
        min: { type: "number", description: "最小值（数字类型）" },
        max: { type: "number", description: "最大值（数字类型）" },
        length: { type: "number", description: "长度（字符串类型）" }
      }
    }
  },
  {
    name: "url_parse",
    description: "解析 URL 获取各部分信息",
    inputSchema: {
      type: "object",
      properties: {
        url: { type: "string", description: "要解析的 URL" }
      },
      required: ["url"]
    }
  },
  {
    name: "timestamp",
    description: "时间戳转换",
    inputSchema: {
      type: "object",
      properties: {
        action: { type: "string", description: "操作类型", enum: ["now", "to_date", "to_timestamp"] },
        value: { type: "string", description: "时间戳或日期字符串" }
      }
    }
  }
];

// 模拟天气数据
const weatherData = {
  '北京': { weather: '晴', temp: 25, humidity: 60 },
  '上海': { weather: '多云', temp: 28, humidity: 75 },
  '广州': { weather: '阵雨', temp: 32, humidity: 85 },
  '深圳': { weather: '晴', temp: 30, humidity: 70 },
  '杭州': { weather: '阴', temp: 26, humidity: 65 },
  '成都': { weather: '小雨', temp: 22, humidity: 80 },
  '武汉': { weather: '晴', temp: 29, humidity: 55 },
  '西安': { weather: '多云', temp: 24, humidity: 50 },
  '南京': { weather: '晴', temp: 27, humidity: 60 },
  '重庆': { weather: '阴', temp: 28, humidity: 75 }
};

// 工具执行函数
async function executeTool(name, args) {
  switch (name) {
    case "get_weather": {
      const city = args.city || '北京';
      const data = weatherData[city] || { weather: '晴', temp: Math.floor(Math.random() * 15) + 20, humidity: Math.floor(Math.random() * 40) + 40 };
      return {
        content: [{ type: "text", text: `${city} 天气: ${data.weather}, 温度: ${data.temp}°C, 湿度: ${data.humidity}%` }]
      };
    }

    case "calculate": {
      try {
        // 安全的数学计算
        const expr = args.expression.replace(/[^0-9+\-*/().%\s]/g, '');
        const result = Function('"use strict"; return (' + expr + ')')();
        return {
          content: [{ type: "text", text: `计算结果: ${args.expression} = ${result}` }]
        };
      } catch (e) {
        return { content: [{ type: "text", text: `计算错误: ${e.message}` }], isError: true };
      }
    }

    case "get_time": {
      const tz = args.timezone || 'Asia/Shanghai';
      try {
        const time = new Date().toLocaleString('zh-CN', { timeZone: tz });
        return {
          content: [{ type: "text", text: `${tz} 当前时间: ${time}` }]
        };
      } catch (e) {
        return {
          content: [{ type: "text", text: `当前时间: ${new Date().toLocaleString('zh-CN')}` }]
        };
      }
    }

    case "system_info": {
      const type = args.type || 'all';
      const info = {};

      if (type === 'all' || type === 'cpu') {
        const cpus = os.cpus();
        info.cpu = {
          model: cpus[0]?.model,
          cores: cpus.length,
          speed: cpus[0]?.speed + ' MHz'
        };
      }
      if (type === 'all' || type === 'memory') {
        info.memory = {
          total: (os.totalmem() / 1024 / 1024 / 1024).toFixed(2) + ' GB',
          free: (os.freemem() / 1024 / 1024 / 1024).toFixed(2) + ' GB',
          used: ((os.totalmem() - os.freemem()) / 1024 / 1024 / 1024).toFixed(2) + ' GB'
        };
      }
      if (type === 'all' || type === 'platform') {
        info.platform = {
          os: os.platform(),
          arch: os.arch(),
          hostname: os.hostname(),
          release: os.release()
        };
      }
      if (type === 'all' || type === 'uptime') {
        const uptime = os.uptime();
        const days = Math.floor(uptime / 86400);
        const hours = Math.floor((uptime % 86400) / 3600);
        const mins = Math.floor((uptime % 3600) / 60);
        info.uptime = `${days}天 ${hours}小时 ${mins}分钟`;
      }

      return {
        content: [{ type: "text", text: JSON.stringify(info, null, 2) }]
      };
    }

    case "dns_lookup": {
      try {
        const recordType = args.type || 'A';
        const records = await dnsResolve(args.domain, recordType);
        return {
          content: [{ type: "text", text: `${args.domain} 的 ${recordType} 记录:\n${records.join('\n')}` }]
        };
      } catch (e) {
        return { content: [{ type: "text", text: `DNS 查询失败: ${e.message}` }], isError: true };
      }
    }

    case "base64": {
      try {
        if (args.action === 'encode') {
          const encoded = Buffer.from(args.text).toString('base64');
          return { content: [{ type: "text", text: `Base64 编码结果:\n${encoded}` }] };
        } else {
          const decoded = Buffer.from(args.text, 'base64').toString('utf8');
          return { content: [{ type: "text", text: `Base64 解码结果:\n${decoded}` }] };
        }
      } catch (e) {
        return { content: [{ type: "text", text: `Base64 处理失败: ${e.message}` }], isError: true };
      }
    }

    case "json_format": {
      try {
        const obj = JSON.parse(args.json);
        const action = args.action || 'format';
        const result = action === 'format' ? JSON.stringify(obj, null, 2) : JSON.stringify(obj);
        return { content: [{ type: "text", text: result }] };
      } catch (e) {
        return { content: [{ type: "text", text: `JSON 处理失败: ${e.message}` }], isError: true };
      }
    }

    case "uuid_generate": {
      const count = Math.min(args.count || 1, 10);
      const uuids = [];
      for (let i = 0; i < count; i++) {
        uuids.push(crypto.randomUUID());
      }
      return { content: [{ type: "text", text: `生成的 UUID:\n${uuids.join('\n')}` }] };
    }

    case "hash": {
      const algorithm = args.algorithm || 'sha256';
      const hash = crypto.createHash(algorithm).update(args.text).digest('hex');
      return { content: [{ type: "text", text: `${algorithm.toUpperCase()} 哈希值:\n${hash}` }] };
    }

    case "random": {
      const type = args.type || 'number';
      if (type === 'number') {
        const min = args.min || 0;
        const max = args.max || 100;
        const num = Math.floor(Math.random() * (max - min + 1)) + min;
        return { content: [{ type: "text", text: `随机数 (${min}-${max}): ${num}` }] };
      } else if (type === 'string') {
        const len = args.length || 16;
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let str = '';
        for (let i = 0; i < len; i++) {
          str += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return { content: [{ type: "text", text: `随机字符串: ${str}` }] };
      } else if (type === 'password') {
        const len = args.length || 16;
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*';
        let str = '';
        for (let i = 0; i < len; i++) {
          str += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return { content: [{ type: "text", text: `随机密码: ${str}` }] };
      }
      return { content: [{ type: "text", text: '未知的随机类型' }], isError: true };
    }

    case "url_parse": {
      try {
        const url = new URL(args.url);
        const info = {
          protocol: url.protocol,
          hostname: url.hostname,
          port: url.port || '(默认)',
          pathname: url.pathname,
          search: url.search,
          hash: url.hash,
          origin: url.origin
        };
        return { content: [{ type: "text", text: `URL 解析结果:\n${JSON.stringify(info, null, 2)}` }] };
      } catch (e) {
        return { content: [{ type: "text", text: `URL 解析失败: ${e.message}` }], isError: true };
      }
    }

    case "timestamp": {
      const action = args.action || 'now';
      if (action === 'now') {
        const now = Date.now();
        return { content: [{ type: "text", text: `当前时间戳: ${now}\n秒级: ${Math.floor(now/1000)}\n日期: ${new Date(now).toLocaleString('zh-CN')}` }] };
      } else if (action === 'to_date') {
        const ts = parseInt(args.value);
        const date = new Date(ts > 9999999999 ? ts : ts * 1000);
        return { content: [{ type: "text", text: `时间戳 ${args.value} 对应日期:\n${date.toLocaleString('zh-CN')}` }] };
      } else if (action === 'to_timestamp') {
        const date = new Date(args.value);
        return { content: [{ type: "text", text: `日期 ${args.value} 对应时间戳:\n毫秒: ${date.getTime()}\n秒: ${Math.floor(date.getTime()/1000)}` }] };
      }
      return { content: [{ type: "text", text: '未知的时间戳操作' }], isError: true };
    }

    default:
      return { content: [{ type: "text", text: `未知工具: ${name}` }], isError: true };
  }
}

const server = http.createServer(async (req, res) => {
  res.setHeader('Content-Type', 'application/json');
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

  if (req.method === 'OPTIONS') {
    res.end();
    return;
  }

  const url = new URL(req.url, `http://localhost:${PORT}`);

  // GET /health
  if (req.method === 'GET' && url.pathname === '/health') {
    res.end(JSON.stringify({ status: 'ok', tools: tools.length }));
    return;
  }

  // GET /tools
  if (req.method === 'GET' && url.pathname === '/tools') {
    res.end(JSON.stringify({ tools }));
    return;
  }

  // POST /tools/{name}
  if (req.method === 'POST' && url.pathname.startsWith('/tools/')) {
    const toolName = url.pathname.split('/')[2];
    let body = '';
    req.on('data', chunk => body += chunk);
    req.on('end', async () => {
      try {
        const { arguments: args } = JSON.parse(body);
        const result = await executeTool(toolName, args || {});
        res.end(JSON.stringify(result));
      } catch (e) {
        res.statusCode = 400;
        res.end(JSON.stringify({ error: e.message }));
      }
    });
    return;
  }

  res.statusCode = 404;
  res.end(JSON.stringify({ error: 'Not found' }));
});

server.listen(PORT, () => {
  console.log(`\n🚀 MCP Server running at http://localhost:${PORT}`);
  console.log(`\n📦 Available tools (${tools.length}):`);
  tools.forEach(t => console.log(`   - ${t.name}: ${t.description}`));
  console.log('\n');
});
