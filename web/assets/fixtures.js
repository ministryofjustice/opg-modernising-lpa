// Account for DOMContentLoaded firing before JS runs
if (document.readyState !== "loading") {
    init()
} else {
    document.addEventListener('DOMContentLoaded', init)
}

function init() {
    const theForm = document.getElementById('the-form');
    const theLink = document.getElementById('the-link');

    function updateTheLink() {
        const data = new FormData(theForm);
        data.delete('csrf');
        const query = Array.from(data).reduce((a, [k, v]) => `${a}&${k}=${v}`, 'redirect=');

        theLink.innerText = `${document.location.origin}${document.location.pathname}?${query}`;
    }

    theForm.addEventListener('change', updateTheLink);
    updateTheLink();
}
