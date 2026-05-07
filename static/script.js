function toggleLike(postId) {
    fetch(`/api/like?id=${postId}`)
        .then(res => res.json())
        .then(data => {
            const countLabel = document.querySelector(`#like-count-${postId}`);
            if (countLabel) countLabel.innerText = data.likes;
        })
        .catch(err => console.error("Erreur API:", err));
}

const createForm = document.getElementById('create-post-form');
if (createForm) {
    createForm.addEventListener('submit', function(e) {
        e.preventDefault();
        const data = Object.fromEntries(new FormData(this));
        fetch('/api/create-post', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        }).then(() => window.location.href = '/forum');
    });
}