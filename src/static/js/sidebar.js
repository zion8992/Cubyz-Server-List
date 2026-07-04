document.addEventListener('DOMContentLoaded', function() {
	const tabs = document.querySelectorAll('.sidebar-tab');
	const panels = document.querySelectorAll('.tab-panel');

	function activateTab(target) {
		const targetTab = document.querySelector('.sidebar-tab[data-tab="' + target + '"]');
		const targetPanel = document.getElementById('tab-' + target);

		if (!targetTab || !targetPanel) return;

		tabs.forEach(function(t) { t.classList.remove('active'); });
		panels.forEach(function(p) { p.classList.remove('active'); });

		targetTab.classList.add('active');
		targetPanel.classList.add('active');
	}

	// Activate tab from URL query parameter
	const params = new URLSearchParams(window.location.search);
	const tabFromUrl = params.get('tab');
	if (tabFromUrl) {
		activateTab(tabFromUrl);
	}

	// Handle tab clicks and update URL
	tabs.forEach(function(tab) {
		tab.addEventListener('click', function() {
			const target = tab.dataset.tab;
			activateTab(target);

			// Update URL without reloading
			const url = new URL(window.location.href);
			url.searchParams.set('tab', target);
			window.history.replaceState({}, '', url);
		});
	});
});
