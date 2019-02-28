let imgs = document.querySelectorAll('img.thumbnail')
for (let i = 0; i < imgs.length; i++){
    let filename = imgs[i].src.split("/").pop()
    if (filename === "dubbzbg.jpg"){
        imgs[i].parentElement.href = "https://www.youtube.com/watch?v=ajlkhFnz8eo"
        imgs[i].parentElement.target= "_blank"
        imgs[i].addEventListener('click', function(){
                alert("I'm gay!")
        })
        break
    }
}
