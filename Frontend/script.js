document.addEventListener('DOMContentLoaded', () => {
    const BASE_URL = 'http://localhost:8000';

    const dropZone = document.getElementById('dropZone');
    const fileInput = document.getElementById('fileInput');
    const fileView = document.getElementById('fileView');
    const fileName = document.getElementById('fileName');
    const btnRemoveFile = document.getElementById('btnRemoveFile');
    
    const tabBtns = document.querySelectorAll('.segment');
    const convertPanel = document.getElementById('convertPanel');
    const compressPanel = document.getElementById('compressPanel');
    
    const btnConvert = document.getElementById('btnConvert');
    const btnCompress = document.getElementById('btnCompress');
    const jpgSettings = document.getElementById('jpgSettings');
    const pngSettings = document.getElementById('pngSettings');
    const jpgSlider = document.getElementById('jpgSlider');
    const pngSlider = document.getElementById('pngSlider');
    const jpgVal = document.getElementById('jpgVal');
    const pngVal = document.getElementById('pngVal');

    const statusArea = document.getElementById('statusArea');
    const loadingState = document.getElementById('loadingState');
    const errorState = document.getElementById('errorState');
    const errorMsg = document.getElementById('errorMsg');
    const successState = document.getElementById('successState');
    const downloadBtn = document.getElementById('downloadBtn');

    // State
    let currentFile = null;
    let fileFormat = null; 

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(evt => {
        dropZone.addEventListener(evt, e => { 
            e.preventDefault(); 
            e.stopPropagation(); 
        });
    });
    
    ['dragenter', 'dragover'].forEach(evt => dropZone.addEventListener(evt, () => dropZone.classList.add('dragover')));
    ['dragleave', 'drop'].forEach(evt => dropZone.addEventListener(evt, () => dropZone.classList.remove('dragover')));

    dropZone.addEventListener('drop', e => handleFiles(e.dataTransfer.files));
    fileInput.addEventListener('change', e => handleFiles(e.target.files));
    btnRemoveFile.addEventListener('click', resetUploader);

    function handleFiles(files) {
        if (!files.length) return;
        const file = files[0];
        
        if (!file.type.match(/image\/(jpeg|jpg|png)/)) {
            return alert("Unsupported file type. Please use JPG or PNG.");
        }

        currentFile = file;
        fileFormat = file.type.includes('png') ? 'png' : 'jpeg';
        
        fileName.textContent = file.name;
        dropZone.classList.add('hidden');
        fileView.classList.remove('hidden');
        resetStatus();
        
        jpgSettings.classList.add('hidden');
        pngSettings.classList.add('hidden');
        if (fileFormat === 'png') {
            pngSettings.classList.remove('hidden');
        } else {
            jpgSettings.classList.remove('hidden');
        }
    }

    function resetUploader() {
        currentFile = null;
        fileInput.value = '';
        dropZone.classList.remove('hidden');
        fileView.classList.add('hidden');
        resetStatus();
    }

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            
            if (btn.dataset.action === 'convert') {
                convertPanel.classList.remove('hidden');
                compressPanel.classList.add('hidden');
            } else {
                convertPanel.classList.add('hidden');
                compressPanel.classList.remove('hidden');
            }
        });
    });

    jpgSlider.addEventListener('input', e => jpgVal.textContent = e.target.value + '%');
    pngSlider.addEventListener('input', e => pngVal.textContent = e.target.value);

    btnConvert.addEventListener('click', () => {
        if (!currentFile) return;
        const formData = new FormData();
        formData.append('upload', currentFile); 
        processRequest('/convert', formData);
    });

    btnCompress.addEventListener('click', () => {
        if (!currentFile) return;
        const formData = new FormData();
        formData.append('upload', currentFile);
        const level = fileFormat === 'png' ? pngSlider.value : jpgSlider.value;
        processRequest(`/compress?level=${level}`, formData);
    });

    async function processRequest(endpoint, formData) {
        fileView.classList.add('hidden');
        resetStatus();
        statusArea.classList.remove('hidden');
        loadingState.classList.remove('hidden');

        try {
            const response = await fetch(`${BASE_URL}${endpoint}`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText || "Server processing failed.");
            }

            const data = await response.json();
            
            if (data && data.file) {
                const downloadUrl = `${BASE_URL}/download?file=${encodeURIComponent(data.file)}`;
                showSuccess(downloadUrl);
            } else {
                throw new Error("Invalid response format from server.");
            }

        } catch (error) {
            loadingState.classList.add('hidden');
            errorState.classList.remove('hidden');
            
            if (error.message === 'Failed to fetch') {
                errorMsg.textContent = "Cannot reach server. Ensure Go backend is running on port 8000.";
            } else {
                errorMsg.textContent = error.message;
            }

            setTimeout(() => {
                resetStatus();
                fileView.classList.remove('hidden');
            }, 3000);
        }
    }

    function resetStatus() {
        statusArea.classList.add('hidden');
        loadingState.classList.add('hidden');
        errorState.classList.add('hidden');
        successState.classList.add('hidden');
    }

    function showSuccess(downloadUrl) {
        loadingState.classList.add('hidden');
        successState.classList.remove('hidden');
        downloadBtn.href = downloadUrl;
    }
});