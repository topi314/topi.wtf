async function loadMoreProjects(after) {
    const button = document.querySelector("#projects-load-more")
    button.disabled = true;
    button.classList.add("loading");

    const response = await fetch(`/api/repositories?after=${after}`, {
        method: "GET"
    });

    if (!response.ok) {
        console.error("error fetching more repositories:", response);
        return;
    }

    const body = await response.text();
    button.remove();
    document.querySelector("#projects").insertAdjacentHTML("beforeend", body);
}

async function loadLastFM() {
    let response;
    try {
        response = await fetch(`/api/lastfm`, {
            method: "GET"
        });
    } catch (e) {
        console.error("error fetching last fm:", e);
        lastFMError();
        return;
    }

    if (!response.ok) {
        console.error("error fetching last fm:", response);
        lastFMError();
        return;
    }

    document.querySelector("#lastfm").innerHTML = await response.text();
}

function lastFMError() {
    document.querySelector("#lastfm").innerHTML = `<span class="error">Error fetching last.fm data</span>`;
}

document.addEventListener('DOMContentLoaded', async () => {
    await loadLastFM();
    setInterval(loadLastFM, 5000);
}, false);
