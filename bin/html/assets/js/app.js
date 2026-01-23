document.addEventListener('DOMContentLoaded', () => {
    // --- DOM 元素引用 --- 
    const dom = {
        currentDate: document.getElementById('current-date'),
        calculateProfitButton: document.getElementById('calculate-profit-button'),
        startDateInput: document.getElementById('start-date-input'),
        endDateInput: document.getElementById('end-date-input'),
        worthInputsContainer: document.getElementById('worth-inputs-container'),
        submitWorthButton: document.getElementById('submit-worth-button'),
        flowTypeSelect: document.getElementById('flow-type-select'),
        profitTableBody: document.getElementById('profit-table-body'),
        allocationTypeSelect: document.getElementById('allocation-type-select'),
        drawAllocationButton: document.getElementById('draw-allocation-button'),
        performanceTypeSelect: document.getElementById('performance-type-select'),
        drawPerformanceButton: document.getElementById('draw-performance-button'),
        flowValueInput: document.getElementById('flow-value-input'),
        submitFlowButton: document.getElementById('submit-flow-button'),
        worthTimeInput: document.getElementById('worth-time-input'),
        flowTimeInput: document.getElementById('flow-time-input'),
        allocationLegend: document.getElementById('allocation-legend'),
    };

    // --- 运行时状态 ---
    let state = {
        assetTypes: [],
        latestProfitData: null,
        charts: {},
        allocationColorMap: {}
    };

    // --- 工具函数 ---
    const { formatNumber, formatDateTimeLocal, hashString, calculateExpression } = window.utils;

    // --- 颜色生成 ---
    const buildAllocationColorMap = (types) => {
        const PALETTE = [
            '#3B82F6', '#1D4ED8', '#38BDF8', '#0EA5E9', '#22C55E', '#16A34A',
            '#F97316', '#FB923C', '#FACC15', '#A855F7', '#7C3AED', '#EC4899',
            '#F43F5E', '#14B8A6', '#0F766E'
        ];
        const map = {};
        const used = new Set();
        (types || []).forEach(t => {
            if (!t || !t.name) return;
            if (t.name === 'cash') {
                map[t.name] = '#CBD5E1';
                return;
            }
            const start = hashString(t.name) % PALETTE.length;
            let picked = null;
            for (let i = 0; i < PALETTE.length; i++) {
                const c = PALETTE[(start + i) % PALETTE.length];
                if (!used.has(c)) {
                    picked = c;
                    break;
                }
            }
            picked = picked || PALETTE[start];
            used.add(picked);
            map[t.name] = picked;
        });
        return map;
    };

    // --- 渲染函数 ---
    const renderAllocationLegend = () => {
        if (!dom.allocationLegend) return;
        dom.allocationLegend.innerHTML = '';
        const { labels, datasets } = state.charts.allocationChart.data;
        const colors = datasets?.[0]?.backgroundColor || [];
        labels.forEach((label, idx) => {
            const item = document.createElement('div');
            item.className = 'flex items-center gap-2 text-neutral';
            item.innerHTML = `
                <span class="w-3 h-3 rounded-full" style="background:${colors[idx] || '#CBD5E1'}"></span>
                <span class="text-sm">${label}</span>
            `;
            dom.allocationLegend.appendChild(item);
        });
    };

    // --- 事件处理 ---
    const handleCalculateProfit = async () => {
        const { startDateInput, endDateInput, profitTableBody, drawAllocationButton, drawPerformanceButton } = dom;
        if (!startDateInput.value || !endDateInput.value) {
            alert('请选择起始和结束日期');
            return;
        }
        profitTableBody.innerHTML = '<tr><td colspan="4" class="py-2 text-center">计算中...</td></tr>';
        try {
            const apiResponse = await api.fetchProfit(startDateInput.value, endDateInput.value);
            if (apiResponse.code === 200 && apiResponse.data) {
                profitTableBody.innerHTML = '';
                state.latestProfitData = apiResponse.data;
                state.assetTypes.forEach(asset => {
                    const assetProfit = state.latestProfitData[asset.name];
                    if (assetProfit && Array.isArray(assetProfit)) {
                        const [currentValue, principal, profit] = assetProfit;
                        const profitClass = profit >= 0 ? 'text-danger' : 'text-success';
                        const row = document.createElement('tr');
                        row.className = 'border-b border-gray-50';
                        row.innerHTML = `
                            <td class="py-2 font-medium">${asset.cname}</td>
                            <td class="py-2">${formatNumber(principal)}</td>
                            <td class="py-2">${formatNumber(currentValue)}</td>
                            <td class="py-2 ${profitClass}">${formatNumber(profit)}</td>
                        `;
                        profitTableBody.appendChild(row);
                    }
                });
                drawAllocationButton.click();
                renderAllocationLegend();
                drawPerformanceButton.click();
            } else {
                profitTableBody.innerHTML = `<tr><td colspan="4" class="py-2 text-center text-red-500">${apiResponse.msg || '加载收益数据失败'}</td></tr>`;
            }
        } catch (error) {
            profitTableBody.innerHTML = '<tr><td colspan="4" class="py-2 text-center text-red-500">请求收益数据错误</td></tr>';
        }
    };

    const handleDrawAllocation = () => {
        if (!state.latestProfitData) return;
        const { allocationTypeSelect } = dom;
        const { allocationChart } = state.charts;
        const typeMap = { worth: 0, principal: 1, profit: 2 };
        const dataIndex = typeMap[allocationTypeSelect.value];
        const labels = [], data = [], colors = [];

        state.assetTypes.forEach(asset => {
            const assetData = state.latestProfitData[asset.name];
            if (assetData && assetData[dataIndex] > 0) {
                labels.push(asset.cname);
                data.push(assetData[dataIndex]);
                colors.push(state.allocationColorMap[asset.name] || '#CBD5E1');
            }
        });

        allocationChart.data.labels = labels;
        allocationChart.data.datasets[0].data = data;
        allocationChart.data.datasets[0].backgroundColor = colors;
        allocationChart.update();
        renderAllocationLegend();
    };

    const handleDrawPerformance = async () => {
        const { startDateInput, endDateInput, performanceTypeSelect } = dom;
        const { performanceChart } = state.charts;
        if (!startDateInput.value || !endDateInput.value) return;
        try {
            const apiResponse = await api.fetchHistory(startDateInput.value, endDateInput.value, performanceTypeSelect.value);
            if (apiResponse.code === 200 && Array.isArray(apiResponse.data)) {
                const historyData = apiResponse.data;
                performanceChart.data.labels = historyData.map(d => new Date(d.time).toLocaleDateString('zh-CN', { year: '2-digit', month: '2-digit', day: '2-digit' }));
                performanceChart.data.datasets[0].data = historyData.map(d => d.worth);
                performanceChart.data.datasets[1].data = historyData.map(d => d.principal);
                performanceChart.data.datasets[2].data = historyData.map(d => d.profit);
                performanceChart.update();
            } else {
                alert(`加载历史数据失败: ${apiResponse.msg || '未知错误'}`);
            }
        } catch (error) {
            alert('请求历史数据错误');
        }
    };

    // 处理输入表达式实时计算
    const handleInputExpression = (input, resultSpan) => {
        const value = input.value.trim();
        const result = calculateExpression(value);
        
        if (value && result !== null) {
            resultSpan.textContent = `= ${formatNumber(result)}`;
            resultSpan.className = 'text-xs text-gray-500 mt-1';
        } else if (value && result === null) {
            resultSpan.textContent = '表达式格式错误';
            resultSpan.className = 'text-xs text-red-400 mt-1';
        } else {
            resultSpan.textContent = '';
            resultSpan.className = '';
        }
    };

    const handleSubmitWorth = async () => {
        const { worthTimeInput, worthInputsContainer } = dom;
        const payload = { time: formatDateTimeLocal(worthTimeInput.value) };
        worthInputsContainer.querySelectorAll('input[type="text"]').forEach(input => {
            let value = 0;
            const expression = input.value.trim();
            
            // 使用统一的表达式计算函数
            const calculated = calculateExpression(expression);
            if (calculated !== null) {
                value = calculated;
            } else {
                // 如果不是表达式，尝试直接解析数字
                value = parseFloat(expression.replace(/,/g, '')) || 0;
            }
            
            payload[input.dataset.name] = isNaN(value) ? 0 : value;
        });
        try {
            const data = await api.submitWorth(payload);
            alert(data.code === 200 ? '提交成功！' : `提交失败: ${data.msg || '未知错误'}`);
        } catch (error) {
            alert('提交请求失败');
        }
    };

    const handleSubmitFlow = async () => {
        const { flowTypeSelect, flowValueInput, flowTimeInput } = dom;
        const value = parseFloat(flowValueInput.value);
        if (!flowTypeSelect.value || isNaN(value)) {
            alert('请选择类型并输入有效的金额');
            return;
        }
        const payload = {
            type: flowTypeSelect.value,
            value: value,
            time: formatDateTimeLocal(flowTimeInput.value)
        };
        try {
            const data = await api.submitFlowRecord(payload);
            if (data.code === 200) {
                alert('资金流动记录提交成功！');
                flowValueInput.value = '';
            } else {
                alert(`提交失败: ${data.msg || '未知错误'}`);
            }
        } catch (error) {
            alert('提交请求失败');
        }
    };

    // --- 初始化 ---
    const init = async () => {
        dom.currentDate.textContent = new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' });

        const today = new Date(), year = today.getFullYear();
        const formatDate = (date) => date.toISOString().split('T')[0];
        dom.startDateInput.value = formatDate(new Date(year, 0, 1));
        dom.endDateInput.value = formatDate(new Date(year, 11, 31));

        try {
            const apiResponse = await api.fetchAssetTypes();
            if (apiResponse.code === 200 && Array.isArray(apiResponse.data)) {
                state.assetTypes = apiResponse.data;
                state.allocationColorMap = buildAllocationColorMap(state.assetTypes);

                // 填充动态内容
                const { worthInputsContainer, flowTypeSelect, performanceTypeSelect } = dom;
                worthInputsContainer.innerHTML = '';
                flowTypeSelect.innerHTML = '';
                performanceTypeSelect.innerHTML = '<option value="all">总体</option>';

                state.assetTypes.forEach(item => {
                    const inputGroup = document.createElement('div');
                    
                    const labelInputWrapper = document.createElement('div');
                    labelInputWrapper.className = 'flex items-center';
                    
                    const label = document.createElement('label');
                    label.className = 'w-20 shrink-0 text-sm font-medium text-neutral';
                    label.textContent = item.cname;
                    
                    const inputWrapper = document.createElement('div');
                    inputWrapper.className = 'flex-1 flex flex-col';
                    
                    const input = document.createElement('input');
                    input.type = 'text';
                    input.dataset.name = item.name;
                    input.className = 'px-3 py-2 border border-gray-200 rounded-lg input-focus text-sm';
                    input.value = '0';
                    
                    const resultSpan = document.createElement('span');
                    resultSpan.className = '';
                    
                    inputWrapper.appendChild(input);
                    inputWrapper.appendChild(resultSpan);
                    
                    labelInputWrapper.appendChild(label);
                    labelInputWrapper.appendChild(inputWrapper);
                    inputGroup.appendChild(labelInputWrapper);
                    
                    // 绑定输入事件，实时计算
                    input.addEventListener('input', () => {
                        handleInputExpression(input, resultSpan);
                    });
                    
                    // 绑定失焦事件，确保最终结果正确显示
                    input.addEventListener('blur', () => {
                        handleInputExpression(input, resultSpan);
                    });
                    
                    worthInputsContainer.appendChild(inputGroup);

                    const perfOption = new Option(item.cname, item.name);
                    performanceTypeSelect.add(perfOption);

                    const flowOption = new Option(item.cname, item.name);
                    flowTypeSelect.add(flowOption);
                });

                // 绑定事件
                dom.calculateProfitButton.addEventListener('click', handleCalculateProfit);
                dom.drawAllocationButton.addEventListener('click', handleDrawAllocation);
                dom.drawPerformanceButton.addEventListener('click', handleDrawPerformance);
                dom.submitWorthButton.addEventListener('click', handleSubmitWorth);
                dom.submitFlowButton.addEventListener('click', handleSubmitFlow);

                // 初始化图表
                state.charts = window.charts.initCharts();

                // 自动加载初始数据
                dom.calculateProfitButton.click();
            } else {
                dom.worthInputsContainer.innerHTML = '<p class="text-red-500">加载资产类型失败</p>';
            }
        } catch (error) {
            dom.worthInputsContainer.innerHTML = '<p class="text-red-500">请求资产类型错误</p>';
        }
    };

    init();
});
