document.addEventListener('DOMContentLoaded', () => {
	const tabs = document.querySelectorAll('.server-tab');
	const panels = document.querySelectorAll('.server-main .tab-panel');

	tabs.forEach(tab => {
		tab.addEventListener('click', () => {
			tabs.forEach(t => t.classList.remove('active'));
			panels.forEach(p => p.classList.remove('active'));

			tab.classList.add('active');
			document.getElementById('tab-' + tab.dataset.tab).classList.add('active');
		});
	});
});
