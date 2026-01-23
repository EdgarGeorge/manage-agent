// 工具函数
const formatNumber = (num) => 
    num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

const formatDateTimeLocal = (value) => {
    if (!value) return '';
    const [date, timeRaw] = value.split('T');
    if (!date || !timeRaw) return value;
    let time = timeRaw.split('.')[0]; // 去掉毫秒
    if (time.length === 5) {
        time = `${time}:00`;
    }
    return `${date} ${time}`;
};

const hashString = (s) => {
    let h = 5381;
    for (let i = 0; i < s.length; i++) 
        h = ((h << 5) + h) + s.charCodeAt(i);
    return Math.abs(h);
};

// 计算加法表达式（安全版本，只支持数字和加号）
const calculateExpression = (expression) => {
    if (!expression || typeof expression !== 'string') return null;
    
    // 移除所有空格
    const cleaned = expression.trim().replace(/\s+/g, '');
    if (!cleaned) return null;
    
    // 验证表达式格式：只允许数字、小数点、加号
    if (!/^[\d.+\s]+$/.test(cleaned)) return null;
    
    try {
        // 分割并计算
        const parts = cleaned.split('+').map(part => {
            const num = parseFloat(part.trim());
            return isNaN(num) ? 0 : num;
        });
        
        const result = parts.reduce((sum, num) => sum + num, 0);
        return isNaN(result) ? null : result;
    } catch (e) {
        return null;
    }
};

// 暴露到全局，以便其他脚本文件可以访问
window.utils = { formatNumber, formatDateTimeLocal, hashString, calculateExpression };
