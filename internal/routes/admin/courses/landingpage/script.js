function updatePreview() {
  document.getElementById('prev-badge').innerText = document.getElementById('input-badge').value;
  document.getElementById('prev-headline').innerText = document.getElementById('input-headline').value;
  document.getElementById('prev-sub').innerText = document.getElementById('input-sub').value;
  document.getElementById('prev-price').innerText = document.getElementById('input-price').value;
  document.getElementById('prev-orig').innerText = document.getElementById('input-orig-price').value;
  document.getElementById('prev-cta').innerText = document.getElementById('input-cta').value;
}

function saveLandingPage() {
  const data = {
    badge: document.getElementById('input-badge').value,
    headline: document.getElementById('input-headline').value,
    sub: document.getElementById('input-sub').value,
    price: document.getElementById('input-price').value,
    origPrice: document.getElementById('input-orig-price').value,
    cta: document.getElementById('input-cta').value
  };
  localStorage.setItem('smg_landing_page_config', JSON.stringify(data));
  alert('Public course landing page saved and live!');
}

function copyLink() {
  navigator.clipboard.writeText('https://salenmobileguru.com/courses/motherboard-repair-diagnostics');
  alert('Public course link copied to clipboard!');
}

// Initial Boot
const saved = JSON.parse(localStorage.getItem('smg_landing_page_config'));
if (saved) {
  document.getElementById('input-badge').value = saved.badge || '';
  document.getElementById('input-headline').value = saved.headline || '';
  document.getElementById('input-sub').value = saved.sub || '';
  document.getElementById('input-price').value = saved.price || '';
  document.getElementById('input-orig-price').value = saved.origPrice || '';
  document.getElementById('input-cta').value = saved.cta || '';
  updatePreview();
}
