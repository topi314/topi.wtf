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

async function loadMorePosts(after) {
    const button = document.querySelector("#posts-load-more")
    button.disabled = true;
    button.classList.add("loading");

    const response = await fetch(`/api/posts?after=${after}`, {
        method: "GET"
    });

    if (!response.ok) {
        console.error("error fetching more posts:", response);
        return;
    }

    const body = await response.text();
    button.remove();
    document.querySelector("#posts").insertAdjacentHTML("beforeend", body);
}