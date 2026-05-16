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
            } else {
                alert("Erreur lors de la création : " + (result.error || "Inconnu"));
            }
        })
        .catch(err => console.error("Erreur API Post:", err));
    });
}

function toggleLike(postId) {
    fetch(`/api/like?id=${postId}`)
        .then(res => res.json())
        .then(data => {
            const countSpan = document.getElementById(`like-count-${postId}`);
            if (countSpan) countSpan.innerText = data.likes;
        })
        .catch(err => console.error("Erreur API Like:", err));
}

const fileInput = document.getElementById('addPP');
let previewImg = document.getElementById('pp-preview');
const cancelBtn = document.getElementById('cancel-pp');

if (fileInput) {
    fileInput.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(ev) {
                if (!previewImg) {
                    previewImg = document.createElement('img');
                    previewImg.id = 'pp-preview';
                    previewImg.style.width = '120px';
                    previewImg.style.height = '120px';
                    previewImg.style.objectFit = 'cover';
                    previewImg.style.borderRadius = '50%';
                    previewImg.style.border = '2px solid #ccc';
                    document.querySelector('#pp-form').parentNode.insertBefore(previewImg, document.querySelector('#pp-form'));
                }
                previewImg.src = ev.target.result;
                previewImg.style.display = 'block';
                if (cancelBtn) cancelBtn.style.display = 'inline-block';
            };
            reader.readAsDataURL(file);
        }
    });
}

if (cancelBtn) {
    cancelBtn.addEventListener('click', function() {
        fileInput.value = '';
        cancelBtn.style.display = 'none';
    });
}
