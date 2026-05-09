const postForm = document.getElementById('create-post-form');
if (postForm) {
    postForm.addEventListener('submit', function(e) {
        e.preventDefault();
        const data = Object.fromEntries(new FormData(this));
        
        fetch('/api/create-post', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        })
        .then(res => res.json())
        .then(result => {
            if (result.status === "success") {
                window.location.href = "/forum";
            }
        });
    });
}

function toggleLike(postId) {
    fetch(`/api/like?id=${postId}`)
        .then(res => res.json())
        .then(data => {
            const countSpan = document.getElementById(`like-count-${postId}`);
            if (countSpan) countSpan.innerText = data.likes;
        });
}