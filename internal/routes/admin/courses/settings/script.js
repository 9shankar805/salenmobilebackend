function updateThumbPreview() {
  document.getElementById('thumb-preview-img').src = document.getElementById('course-thumb-url').value;
}

function saveSettings() {
  const data = {
    title: document.getElementById('course-title').value,
    slug: document.getElementById('course-slug').value,
    summary: document.getElementById('course-summary').value,
    category: document.getElementById('course-category').value,
    level: document.getElementById('course-level').value,
    thumb: document.getElementById('course-thumb-url').value,
    price: document.getElementById('course-price').value,
    origPrice: document.getElementById('course-orig-price').value,
    selfEnroll: document.getElementById('course-self-enroll').checked
  };
  localStorage.setItem('smg_active_course_settings', JSON.stringify(data));
  alert('Course settings successfully saved!');
}

function deleteCourse() {
  if (confirm('Are you ABSOLUTELY sure you want to delete this course? This action cannot be undone.')) {
    alert('Course marked for deletion.');
    window.location.href = '/admin/courses';
  }
}
