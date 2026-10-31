// Chart 1: Growth
new Chart(document.getElementById('growthChart'), {
  type: 'line',
  data: {
    labels: ['Oct 1', 'Oct 5', 'Oct 10', 'Oct 15', 'Oct 20', 'Oct 25', 'Oct 30'],
    datasets: [{
      label: 'Active Students',
      data: [8, 11, 14, 15, 17, 18, 18],
      borderColor: '#6366f1',
      backgroundColor: 'rgba(99,102,241,0.1)',
      fill: true,
      tension: 0.3
    }]
  },
  options: {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { grid: { color: 'rgba(255,255,255,0.05)' } },
      y: { grid: { color: 'rgba(255,255,255,0.05)' }, min: 0 }
    }
  }
});

// Chart 2: Funnel
new Chart(document.getElementById('funnelChart'), {
  type: 'bar',
  data: {
    labels: ['Mod 1: Power', 'Mod 2: BGA', 'Mod 3: Audio', 'Mod 4: Baseband'],
    datasets: [{
      label: 'Completions',
      data: [18, 16, 12, 9],
      backgroundColor: ['#10b981', '#6366f1', '#06b6d4', '#8b5cf6'],
      borderRadius: 6
    }]
  },
  options: {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { grid: { color: 'rgba(255,255,255,0.05)' } },
      y: { grid: { color: 'rgba(255,255,255,0.05)' }, min: 0, max: 20 }
    }
  }
});

function exportCSV() {
  const csv = 'Student,Progress %,Avg Lab Score %,Last Active,Status\n' +
    'Aarav Sharma,94%,96.5%,2 hours ago,High Performer\n' +
    'Bikash Gurung,88%,91.0%,5 hours ago,High Performer\n' +
    'Chandra Prakash Thapa,75%,82.5%,Yesterday,Steady\n' +
    'Dipesh Shrestha,68%,78.0%,2 days ago,Steady\n' +
    'Gopal Adhikari,62%,71.0%,4 days ago,Needs Practice\n';

  const blob = new Blob([csv], { type: 'text/csv' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = 'course_analytics.csv';
  a.click();
}
