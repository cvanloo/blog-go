window.addEventListener('DOMContentLoaded', () => {
  const observer = new IntersectionObserver(entries => {
    entries.forEach(entry => {
      const id = entry.target.getAttribute('id');
      if (entry.intersectionRatio > 0) {
        document.querySelector(`ul li a[href="#${id}"]`)?.classList.add('active');
      } else {
        document.querySelector(`ul li a[href="#${id}"]`)?.classList.remove('active');
      }
    });
  });
  document.querySelectorAll('section[id]').forEach(section => {
    observer.observe(section);
  });
});
