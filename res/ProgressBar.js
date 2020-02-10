function UploadProgressBar(uploadManager, queueDiv, file){
	this.uploadManager = uploadManager
	this.file = file
	this.name = file.name

	this.uploadDiv = document.createElement("div")
	this.uploadDiv.classList.add("file_button")
	this.uploadDiv.style.opacity = "0"
	this.uploadDiv.innerText = "Queued\n" + this.file.name
	queueDiv.appendChild(this.uploadDiv)

	// Start uploading the file
	this.uploadManager.addFile(
		this.file,
		this.name,
		(progress) => { this.onProgress(progress) },
		(id)       => { this.onFinished(id) },
		(val, msg) => { this.onFailure(val, msg) }
	)

	// Browsers don't render the transition if the opacity is set and
	// updated in the same frame. So we have to wait a frame (or more)
	// before changing the opacity to make sure the transition triggers
	window.setTimeout(() => {this.uploadDiv.style.opacity = "1"}, 100)
}

UploadProgressBar.prototype.onProgress = function(progress){
	this.uploadDiv.innerText = "Uploading... " + Math.round(progress*1000)/10 + "%\n" + this.name
	this.uploadDiv.style.background = 'linear-gradient('
		+'to right, '
		+'var(--file_background_color) 0%, '
		+'var(--highlight_color) '+ ((progress*100)) +'%, '
		+'var(--file_background_color) '+ ((progress*100)+1) +'%)'
}
UploadProgressBar.prototype.onFinished = function(id){
	console.log("Upload finished: "+this.file.name+" "+id)
	let link = "sia://"+id

	this.uploadDiv.style.background = 'var(--file_background_color)'

	let linkNode = document.createElement("a")
	linkNode.href = "/"+id
	linkNode.target = "_blank"
	linkNode.innerText = link+"\n"+this.file.name

	let copyBtn = document.createElement("img")
	copyBtn.src = "/res/copy.svg"
	copyBtn.addEventListener("click", e => {
		this.copyLink(link)
	})

	this.uploadDiv.innerHTML = ""
	this.uploadDiv.appendChild(linkNode)
	this.uploadDiv.appendChild(copyBtn)

	// Add this file to the upload history
	this.addHistory(link, this.file.name)
}
UploadProgressBar.prototype.onFailure = function(val, msg) {
	if (val === "") {
		val = "Could not connect to server"
	}

	this.uploadDiv.innerHTML = "" // Remove uploading progress
	this.uploadDiv.style.background = 'var(--danger_color)'
	this.uploadDiv.style.color = 'var(--highlight_text_color)'
	this.uploadDiv.innerText = "Error: "+msg+" ("+val+")\n"+this.file.name
	console.log(msg)
}
UploadProgressBar.prototype.copyLink = function(text) {
	let node = document.createElement("span")
	if (copyLink(text)) {
		node.innerText = "Link copied!\n"
	} else {
		node.innerText = "Copy failed, please copy link manually!\n"
	}
	this.uploadDiv.prepend(node)

	window.setTimeout(() => {
		node.remove()
	}, 10000)
}
UploadProgressBar.prototype.addHistory = function(link, name) {
	let uploadsStr = localStorage.getItem("uploaded_files");
	if (uploadsStr === null) { uploadsStr = "[]"; }

	let uploads = []
	try {
		uploads = JSON.parse(uploadsStr)
	} catch (err) {
		uploads = []
	}

	// Check if there are not too many values stored
	if (uploads.length > 1000) {
		uploads.pop()
		uploads.pop()
	}

	// Prepend the item
	uploads.unshift({link: link, name: name})

	// Save the new file
	localStorage.setItem("uploaded_files", JSON.stringify(uploads));
}
