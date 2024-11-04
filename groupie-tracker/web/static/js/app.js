// web/static/js/app.js

// Constants and Globals
const MAPBOX_TOKEN = 'pk.eyJ1Ijoic3RlbGxhYWNoYXJvaXJvIiwiYSI6ImNtMWhmZHNlODBlc3cybHF5OWh1MDI2dzMifQ.wk3v-v7IuiSiPwyq13qdHw';
let allArtists = [];
let allLocations = new Set();

// DOM Elements
const elements = {
    searchInput: document.getElementById('search-input'),
    suggestionsContainer: document.getElementById('suggestions'),
    creationYearSlider: document.getElementById('creation-year'),
    creationYearDisplay: document.getElementById('creation-year-display'),
    firstAlbumYearSlider: document.getElementById('first-album-year'),
    firstAlbumYearDisplay: document.getElementById('first-album-year-display'),
    memberCheckboxes: document.getElementById('member-checkboxes'),
    locationCheckboxes: document.getElementById('location-checkboxes'),
    resultsContainer: document.getElementById('results-container'),
    loading: document.getElementById('loading'),
    errorMessage: document.getElementById('error-message')
};

// Utility Functions
function showLoading() {
    elements.loading.style.display = 'flex';
}

function hideLoading() {
    elements.loading.style.display = 'none';
}

function showError(message) {
    elements.errorMessage.textContent = message;
    elements.errorMessage.style.display = 'block';
    setTimeout(() => {
        elements.errorMessage.style.display = 'none';
    }, 5000);
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Filter Functions
function updateCreationYearDisplay() {
    elements.creationYearDisplay.textContent = elements.creationYearSlider.value;
    applyFilters();
}

function updateFirstAlbumYearDisplay() {
    elements.firstAlbumYearDisplay.textContent = elements.firstAlbumYearSlider.value;
    applyFilters();
}

function getFilterValues() {
    return {
        creationYearMin: parseInt(elements.creationYearSlider.value),
        creationYearMax: 2023,
        firstAlbumYearMin: parseInt(elements.firstAlbumYearSlider.value),
        firstAlbumYearMax: 2023,
        members: Array.from(elements.memberCheckboxes.querySelectorAll('input:checked'))
            .map(cb => parseInt(cb.value)),
        locations: Array.from(elements.locationCheckboxes.querySelectorAll('input:checked'))
            .map(cb => cb.value)
    };
}

function initializeFilters() {
    // Set up creation year slider
    const earliestCreationYear = Math.min(...allArtists.map(artist => artist.creationDate));
    const latestCreationYear = Math.max(...allArtists.map(artist => artist.creationDate));
    elements.creationYearSlider.min = earliestCreationYear;
    elements.creationYearSlider.max = latestCreationYear;
    elements.creationYearSlider.value = earliestCreationYear;
    elements.creationYearDisplay.textContent = earliestCreationYear;

    // Set up first album year slider
    const earliestAlbumYear = Math.min(...allArtists.map(artist => 
        parseInt(artist.firstAlbum.split('-')[2])));
    const latestAlbumYear = Math.max(...allArtists.map(artist => 
        parseInt(artist.firstAlbum.split('-')[2])));
    elements.firstAlbumYearSlider.min = earliestAlbumYear;
    elements.firstAlbumYearSlider.max = latestAlbumYear;
    elements.firstAlbumYearSlider.value = earliestAlbumYear;
    elements.firstAlbumYearDisplay.textContent = earliestAlbumYear;

    setupMemberCheckboxes();
    setupLocationCheckboxes();
}

function setupMemberCheckboxes() {
    const maxMembers = Math.max(...allArtists.map(artist => artist.members.length));
    elements.memberCheckboxes.innerHTML = '';
    
    for (let i = 1; i <= maxMembers; i++) {
        const label = document.createElement('label');
        label.innerHTML = `<input type="checkbox" value="${i}"> ${i}`;
        elements.memberCheckboxes.appendChild(label);
    }
    
    elements.memberCheckboxes.querySelectorAll('input').forEach(checkbox => {
        checkbox.addEventListener('change', applyFilters);
    });
}

function setupLocationCheckboxes() {
    allLocations = new Set();
    allArtists.forEach(artist => {
        artist.locations.forEach(location => {
            allLocations.add(location.split('-')[0].trim());
        });
    });

    elements.locationCheckboxes.innerHTML = '';
    Array.from(allLocations).sort().forEach(location => {
        const label = document.createElement('label');
        label.innerHTML = `<input type="checkbox" value="${location}"> ${location}`;
        elements.locationCheckboxes.appendChild(label);
    });

    elements.locationCheckboxes.querySelectorAll('input').forEach(checkbox => {
        checkbox.addEventListener('change', applyFilters);
    });
}

// Search and Display Functions
function displaySuggestions(suggestions) {
    elements.suggestionsContainer.innerHTML = '';
    suggestions.forEach(suggestion => {
        const div = document.createElement('div');
        div.className = 'suggestion-item';
        div.textContent = `${suggestion.text} (${suggestion.type})`;
        div.onclick = () => {
            elements.searchInput.value = suggestion.text;
            elements.suggestionsContainer.innerHTML = '';
            searchArtists(suggestion.text);
        };
        elements.suggestionsContainer.appendChild(div);
    });
}

async function searchArtists(query) {
    showLoading();
    const filters = getFilterValues();

    try {
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(filters)
        });

        if (!response.ok) {
            throw new Error('Network response was not ok');
        }

        const data = await response.json();
        displayResults(data.artists);

        if (!allArtists.length) {
            allArtists = data.artists;
            initializeFilters();
        }
    } catch (error) {
        console.error('Error:', error);
        showError('An error occurred while searching for artists. Please try again later.');
    } finally {
        hideLoading();
    }
}

function displayResults(artists) {
    elements.resultsContainer.innerHTML = '';
    
    if (artists.length === 0) {
        elements.resultsContainer.innerHTML = `
            <div class="no-results">
                <p>No artists found matching your criteria.</p>
            </div>
        `;
        return;
    }

    artists.forEach(artist => {
        const card = createArtistCard(artist);
        elements.resultsContainer.appendChild(card);
    });

    lazyLoadImages();
}

function createArtistCard(artist) {
    const card = document.createElement('div');
    card.className = 'artist-card';
    card.innerHTML = `
        <img src="placeholder.jpg" data-src="${artist.image}" alt="${artist.name}" class="lazy-image">
        <h3>${artist.name}</h3>
        <p><i class="fas fa-calendar-alt"></i> Created: ${artist.creationDate}</p>
        <p><i class="fas fa-compact-disc"></i> First Album: ${artist.firstAlbum}</p>
    `;
    card.onclick = () => {
        window.location.href = `/artist/${artist.id}`;
    };
    return card;
}

function lazyLoadImages() {
    const options = {
        root: null,
        rootMargin: '0px',
        threshold: 0.1
    };

    const observer = new IntersectionObserver((entries, observer) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const img = entry.target;
                img.src = img.dataset.src;
                img.classList.remove('lazy-image');
                observer.unobserve(img);
            }
        });
    }, options);

    document.querySelectorAll('.lazy-image').forEach(img => observer.observe(img));
}

// Event Listeners
elements.searchInput.addEventListener('input', debounce(() => {
    const query = elements.searchInput.value;
    if (query.length >= 1) {
        fetch(`/api/suggestions?q=${encodeURIComponent(query)}`)
            .then(response => response.json())
            .then(suggestions => displaySuggestions(suggestions))
            .catch(error => console.error('Error:', error));
    } else {
        elements.suggestionsContainer.innerHTML = '';
    }
}, 300));

elements.creationYearSlider.addEventListener('input', updateCreationYearDisplay);
elements.firstAlbumYearSlider.addEventListener('input', updateFirstAlbumYearDisplay);

// Initialize
window.addEventListener('load', () => {
    searchArtists('');
});