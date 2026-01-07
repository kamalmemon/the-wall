document.addEventListener('DOMContentLoaded', function() {
  const form = document.getElementById('entryForm');
  const nameInput = document.getElementById('nameInput');
  const nameCharCount = document.getElementById('nameCharCount');
  const messageInput = document.getElementById('messageInput');
  const charCount = document.getElementById('charCount');
  const submitBtn = document.getElementById('submitBtn');
  const guestbook = document.getElementById('guestbook');
  const emptyState = document.getElementById('emptyState');
  const heatmapPicker = document.getElementById('heatmapPicker');
  const pickerCursor = document.getElementById('pickerCursor');
  const colorPreview = document.getElementById('colorPreview');
  const colorValue = document.getElementById('colorValue');

  let currentColor = window.VISITOR_COLOR || '#6B8E9B';

  // Initialize
  colorPreview.style.background = currentColor;
  colorValue.value = currentColor;

  // Color picker
  function getColorFromPosition(x, width) {
    const percentage = x / width;
    const colors = [
      { pos: 0, r: 255, g: 107, b: 107 },
      { pos: 0.15, r: 255, g: 167, b: 38 },
      { pos: 0.30, r: 255, g: 238, b: 88 },
      { pos: 0.45, r: 102, g: 187, b: 106 },
      { pos: 0.60, r: 66, g: 165, b: 245 },
      { pos: 0.75, r: 126, g: 87, b: 194 },
      { pos: 0.90, r: 236, g: 64, b: 122 },
      { pos: 1, r: 255, g: 107, b: 107 },
    ];

    let lower = colors[0], upper = colors[1];
    for (let i = 0; i < colors.length - 1; i++) {
      if (percentage >= colors[i].pos && percentage <= colors[i + 1].pos) {
        lower = colors[i];
        upper = colors[i + 1];
        break;
      }
    }

    const range = upper.pos - lower.pos;
    const localPos = range > 0 ? (percentage - lower.pos) / range : 0;
    const r = Math.round(lower.r + (upper.r - lower.r) * localPos);
    const g = Math.round(lower.g + (upper.g - lower.g) * localPos);
    const b = Math.round(lower.b + (upper.b - lower.b) * localPos);

    return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
  }

  function updatePickerFromPosition(x) {
    const rect = heatmapPicker.getBoundingClientRect();
    const clampedX = Math.max(0, Math.min(x - rect.left, rect.width));
    const percentage = clampedX / rect.width;
    
    pickerCursor.style.left = `${percentage * 100}%`;
    currentColor = getColorFromPosition(clampedX, rect.width);
    colorPreview.style.background = currentColor;
    colorValue.value = currentColor;
  }

  let isDragging = false;

  heatmapPicker.addEventListener('mousedown', (e) => {
    isDragging = true;
    updatePickerFromPosition(e.clientX);
  });

  document.addEventListener('mousemove', (e) => {
    if (isDragging) updatePickerFromPosition(e.clientX);
  });

  document.addEventListener('mouseup', () => isDragging = false);

  heatmapPicker.addEventListener('touchstart', (e) => {
    isDragging = true;
    updatePickerFromPosition(e.touches[0].clientX);
  });

  document.addEventListener('touchmove', (e) => {
    if (isDragging) updatePickerFromPosition(e.touches[0].clientX);
  });

  document.addEventListener('touchend', () => isDragging = false);

  // Character counters
  nameInput.addEventListener('input', function() {
    nameCharCount.textContent = this.value.length;
  });

  messageInput.addEventListener('input', function() {
    charCount.textContent = this.value.length;
  });

  // Form submission
  form.addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const message = messageInput.value.trim();
    if (!message) {
      messageInput.focus();
      return;
    }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Posting...';

    try {
      const name = nameInput.value.trim() || `Visitor #${window.VISITOR_NUMBER}`;
      
      const response = await fetch('/api/entry', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ color: currentColor, name, message }),
      });

      if (!response.ok) throw new Error('Failed');

      const entry = await response.json();
      
      emptyState.style.display = 'none';
      
      const entryEl = document.createElement('div');
      entryEl.className = 'entry';
      entryEl.style.background = entry.color;
      entryEl.innerHTML = `
        <span class="entry-name">${escapeHtml(entry.name)}</span>
        <p class="entry-message">${escapeHtml(entry.message)}</p>
      `;
      guestbook.prepend(entryEl);
      
      messageInput.value = '';
      nameInput.value = '';
      charCount.textContent = '0';
      nameCharCount.textContent = '0';
      
    } catch (error) {
      console.error('Error:', error);
      alert('Failed to post. Please try again.');
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Post';
    }
  });

  function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  pickerCursor.style.left = '50%';
});
