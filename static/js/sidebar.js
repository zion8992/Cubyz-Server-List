document.addEventListener('DOMContentLoaded', function() {
	const tabs = document.querySelectorAll('.sidebar-tab');
	const panels = document.querySelectorAll('.tab-panel');

	tabs.forEach(function(tab) {
		tab.addEventListener('click', function() {
			const target = tab.dataset.tab;

			tabs.forEach(function(t) { t.classList.remove('active'); });
			panels.forEach(function(p) { p.classList.remove('active'); });

			tab.classList.add('active');
			document.getElementById('tab-' + target).classList.add('active');
		});
	});
});
