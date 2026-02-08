const API_BASE = '/api/v1/finance';

const api = {
    // 获取资产类型（使用 enum 接口）
    fetchAssetTypes: async () => {
        const res = await fetch(`${API_BASE}/enum/`);
        return res.json();
    },

    // 提交当前市值
    submitWorth: async (data) => {
        const res = await fetch(`${API_BASE}/worth/`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        return res.json();
    },

    // 提交资金流动记录
    submitFlowRecord: async (data) => {
        const res = await fetch(`${API_BASE}/flow-record/`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        return res.json();
    },

    // 获取收益数据
    fetchProfit: async (startDate, endDate) => {
        const url = `${API_BASE}/profit/?start_date=${startDate}&end_date=${endDate}`;
        const res = await fetch(url);
        return res.json();
    },

    // 获取历史数据
    fetchHistory: async (startDate, endDate, type) => {
        const url = `${API_BASE}/profit/history/?start_date=${startDate}&end_date=${endDate}&type=${type}`;
        const res = await fetch(url);
        return res.json();
    },

    // 资金类型管理相关接口
    // 获取所有资金类型（使用 enum 接口，与 fetchAssetTypes 相同）
    fetchTypes: async () => {
        const res = await fetch(`${API_BASE}/enum/`);
        return res.json();
    },

    // 创建资金类型
    createType: async (data) => {
        const res = await fetch(`${API_BASE}/ftype/`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        return res.json();
    },

    // 删除资金类型
    deleteType: async (id) => {
        const res = await fetch(`${API_BASE}/ftype/${id}`, {
            method: 'DELETE'
        });
        return res.json();
    },

    // 市值记录查询与删除
    fetchWorthHistory: async (pageIndex = 1, pageLimit = 10) => {
        const res = await fetch(`${API_BASE}/worth/?page_index=${pageIndex}&page_limit=${pageLimit}`);
        return res.json();
    },

    deleteWorthRecord: async (id) => {
        const res = await fetch(`${API_BASE}/worth/${id}`, {
            method: 'DELETE'
        });
        return res.json();
    },

    // 流水记录查询与删除
    fetchFlowHistory: async (pageIndex = 1, pageLimit = 10) => {
        const res = await fetch(`${API_BASE}/flow-record/?page_index=${pageIndex}&page_limit=${pageLimit}`);
        return res.json();
    },

    deleteFlowRecord: async (id) => {
        const res = await fetch(`${API_BASE}/flow-record/${id}`, {
            method: 'DELETE'
        });
        return res.json();
    }
};

window.api = api;
