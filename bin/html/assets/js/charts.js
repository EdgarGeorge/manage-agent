// 初始化所有图表
const initCharts = () => {
    // 初始化变化曲线图表
    const performanceCtx = document.getElementById('performanceChart').getContext('2d');
    const performanceChart = new Chart(performanceCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [
                {
                    label: '现值',
                    data: [],
                    borderColor: '#EF4444',
                    backgroundColor: 'rgba(239, 68, 68, 0.1)',
                    borderWidth: 3,
                    tension: 0.3,
                    fill: false,
                    pointBackgroundColor: '#EF4444',
                    pointRadius: 4
                },
                {
                    label: '本金',
                    data: [],
                    borderColor: '#1E293B',
                    backgroundColor: 'rgba(30, 41, 59, 0.1)',
                    borderWidth: 3,
                    tension: 0.3,
                    fill: false,
                    pointBackgroundColor: '#1E293B',
                    pointRadius: 4
                },
                {
                    label: '收益',
                    data: [],
                    borderColor: '#F59E0B',
                    backgroundColor: 'rgba(245, 158, 11, 0.1)',
                    borderWidth: 3,
                    tension: 0.3,
                    fill: false,
                    pointBackgroundColor: '#F59E0B',
                    pointRadius: 4
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'top',
                    labels: {
                        usePointStyle: true,
                        boxWidth: 6
                    }
                },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    titleColor: '#1E293B',
                    bodyColor: '#64748B',
                    borderColor: 'rgba(0, 0, 0, 0.1)',
                    borderWidth: 1,
                    padding: 10,
                    boxPadding: 5,
                    usePointStyle: true,
                    callbacks: {
                        label: function (context) {
                            return `${context.dataset.label}: ${context.raw.toLocaleString('zh-CN')}`;
                        }
                    }
                }
            },
            scales: {
                x: { grid: { display: false } },
                y: {
                    beginAtZero: true,
                    ticks: { callback: value => value.toLocaleString('zh-CN') }
                }
            },
            interaction: { mode: 'nearest', axis: 'x', intersect: false }
        }
    });

    // 初始化资产配比图表
    const allocationCtx = document.getElementById('allocationChart').getContext('2d');
    const allocationChart = new Chart(allocationCtx, {
        type: 'doughnut',
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [],
                borderWidth: 0,
                hoverOffset: 5
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            cutout: '70%',
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    titleColor: '#1E293B',
                    bodyColor: '#64748B',
                    borderColor: 'rgba(0, 0, 0, 0.1)',
                    borderWidth: 1,
                    padding: 10,
                    callbacks: {
                        label: function (context) {
                            const raw = Number(context.raw) || 0;
                            const dataArr = context.dataset.data;
                            const total = dataArr.reduce((sum, v) => sum + (Number(v) || 0), 0);
                            const pct = total > 0 ? (raw / total) * 100 : 0;
                            return `${context.label}: ${raw.toLocaleString('zh-CN')}（${pct.toFixed(2)}%）`;
                        }
                    }
                }
            }
        }
    });

    return { performanceChart, allocationChart };
};

window.charts = { initCharts };
