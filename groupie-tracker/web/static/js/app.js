
// Constants and API
const MAPBOX_TOKEN = 'pk.eyJ1Ijoic3RlbGxhYWNoYXJvaXJvIiwiYSI6ImNtMWhmZHNlODBlc3cybHF5OWh1MDI2dzMifQ.wk3v-v7IuiSiPwyq13qdHw';
mapboxgl.accessToken = MAPBOX_TOKEN;  

// Constants and API
const API = {
    async search(query, filters) {
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(filters)
        });
        if (!response.ok) throw new Error('Failed to search artists');
        return response.json();
    },

    async getSuggestions(query) {
        const response = await fetch(`/api/suggestions?q=${encodeURIComponent(query)}`);
        if (!response.ok) throw new Error('Failed to get suggestions');
        return response.json();
    }
};

// Filter Manager Class
class FilterManager {
    constructor(elements) {
        this.elements = elements;
        this.currentYear = new Date().getFullYear();
    }

    getValues() {
        return {
            creationYearMin: parseInt(this.elements.creationYearSlider.value),
            creationYearMax: this.currentYear,
            firstAlbumYearMin: parseInt(this.elements.firstAlbumYearSlider.value),
            firstAlbumYearMax: this.currentYear,
            members: Array.from(this.elements.memberCheckboxes.querySelectorAll('input:checked'))
                .map(cb => parseInt(cb.value)),
            locations: Array.from(this.elements.locationCheckboxes.querySelectorAll('input:checked'))
                .map(cb => cb.value)
        };
    }

    initialize(data) {
        if (!data || !data.length) return;

        // Setup year sliders
        const creationYears = data.map(artist => artist.creationDate);
        const albumYears = data.map(artist => parseInt(artist.firstAlbum.split('-')[2]));

        this.setupYearSlider(
            this.elements.creationYearSlider,
            this.elements.creationYearDisplay,
            Math.min(...creationYears),
            Math.max(...creationYears)
        );

        this.setupYearSlider(
            this.elements.firstAlbumYearSlider,
            this.elements.firstAlbumYearDisplay,
            Math.min(...albumYears),
            Math.max(...albumYears)
        );

        // Setup filters
        this.setupMemberCheckboxes(data);
        this.setupLocationCheckboxes(data);
    }

    setupYearSlider(slider, display, min, max) {
        slider.min = min;
        slider.max = max;
        slider.value = min;
        display.textContent = min;
    }

    setupMemberCheckboxes(artists) {
        const maxMembers = Math.max(...artists.map(artist => artist.members.length));
        this.elements.memberCheckboxes.innerHTML = '';
        
        for (let i = 1; i <= maxMembers; i++) {
            const label = document.createElement('label');
            label.innerHTML = `<input type="checkbox" value="${i}"> ${i}`;
            this.elements.memberCheckboxes.appendChild(label);
        }
        
        this.addFilterChangeListeners(this.elements.memberCheckboxes);
    }

    setupLocationCheckboxes(artists) {
        const locationSet = new Set();

        // Get unique locations from the LocationsData
        artists.forEach(artist => {
            if (artist.locations && Array.isArray(artist.locations)) {
                artist.locations.forEach(location => {
                    const city = location.split('-')[0].trim();
                    if (city) locationSet.add(city);
                });
            }
        });

        // Create checkboxes
        this.elements.locationCheckboxes.innerHTML = '';
        Array.from(locationSet).sort().forEach(location => {
            const label = document.createElement('label');
            label.innerHTML = `<input type="checkbox" value="${location}"> ${location}`;
            this.elements.locationCheckboxes.appendChild(label);
        });

        this.addFilterChangeListeners(this.elements.locationCheckboxes);
    }

    addFilterChangeListeners(container) {
        container.querySelectorAll('input').forEach(input => {
            input.addEventListener('change', () => this.onFilterChange());
        });
    }

    onFilterChange() {
        if (this.changeCallback) this.changeCallback();
    }
}

// Main App Class
class App {
    constructor() {
        this.elements = {
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

        this.filterManager = new FilterManager(this.elements);
        this.filterManager.changeCallback = () => this.searchArtists(this.elements.searchInput.value);
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Search input with debounce
        this.elements.searchInput.addEventListener('input', 
            this.debounce(() => this.handleSearchInput(), 300)
        );

        // Year slider inputs
        this.elements.creationYearSlider.addEventListener('input', 
            () => this.handleYearInput('creation')
        );
        this.elements.firstAlbumYearSlider.addEventListener('input', 
            () => this.handleYearInput('firstAlbum')
        );

        // Close suggestions on outside click
        document.addEventListener('click', (e) => {
            if (!this.elements.searchInput.contains(e.target) && 
                !this.elements.suggestionsContainer.contains(e.target)) {
                this.elements.suggestionsContainer.innerHTML = '';
            }
        });
    }

    // UI Helpers
    showLoading() {
        this.elements.loading.style.display = 'flex';
    }

    hideLoading() {
        this.elements.loading.style.display = 'none';
    }

    showError(message) {
        this.elements.errorMessage.textContent = message;
        this.elements.errorMessage.style.display = 'block';
        setTimeout(() => {
            this.elements.errorMessage.style.display = 'none';
        }, 5000);
    }

    // Event Handlers
    async handleSearchInput() {
        const query = this.elements.searchInput.value.trim();
        if (query.length >= 1) {
            try {
                const suggestions = await API.getSuggestions(query);
                this.displaySuggestions(suggestions);
            } catch (error) {
                console.error('Error:', error);
                this.elements.suggestionsContainer.innerHTML = '';
            }
        } else {
            this.elements.suggestionsContainer.innerHTML = '';
        }
    }

    handleYearInput(type) {
        const slider = this.elements[`${type}YearSlider`];
        const display = this.elements[`${type}YearDisplay`];
        display.textContent = slider.value;
        this.searchArtists(this.elements.searchInput.value);
    }

    // Search and Display
    async searchArtists(query) {
        this.showLoading();
        try {
            const data = await API.search(query, this.filterManager.getValues());
            this.displayResults(data.artists);
            
            // Initialize filters if first time
            if (!this.initialized) {
                this.filterManager.initialize(data.artists);
                this.initialized = true;
            }
        } catch (error) {
            console.error('Error:', error);
            this.showError('Failed to search artists');
        } finally {
            this.hideLoading();
        }
    }

    displaySuggestions(suggestions) {
        this.elements.suggestionsContainer.innerHTML = '';
        
        if (!suggestions || !suggestions.length) {
            this.elements.suggestionsContainer.style.display = 'none';
            return;
        }

        suggestions.forEach(suggestion => {
            const div = document.createElement('div');
            div.className = 'suggestion-item';
            div.textContent = `${suggestion.text} (${suggestion.type})`;
            div.onclick = () => {
                this.elements.searchInput.value = suggestion.text;
                this.elements.suggestionsContainer.innerHTML = '';
                this.searchArtists(suggestion.text);
            };
            this.elements.suggestionsContainer.appendChild(div);
        });

        this.elements.suggestionsContainer.style.display = 'block';
    }

    displayResults(artists) {
        this.elements.resultsContainer.innerHTML = '';
        
        if (!artists || !artists.length) {
            this.elements.resultsContainer.innerHTML = `
                <div class="no-results">
                    <p>No artists found matching your criteria.</p>
                </div>
            `;
            return;
        }

        artists.forEach(artist => {
            this.elements.resultsContainer.appendChild(this.createArtistCard(artist));
        });

        this.initializeLazyLoading();
    }

    createArtistCard(artist) {
        const card = document.createElement('div');
        card.className = 'artist-card';
        card.innerHTML = `
            <img src="placeholder.jpg" data-src="${artist.image}" alt="${artist.name}" class="lazy-image">
            <h3>${artist.name}</h3>
            <p><i class="fas fa-calendar-alt"></i> Created: ${artist.creationDate}</p>
            <p><i class="fas fa-compact-disc"></i> First Album: ${artist.firstAlbum}</p>
        `;
        card.onclick = () => window.location.href = `/artist/${artist.id}`;
        return card;
    }

    // Utility Functions
    debounce(func, wait) {
        let timeout;
        return (...args) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }

    initializeLazyLoading() {
        const observer = new IntersectionObserver(
            (entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        img.src = img.dataset.src;
                        img.classList.remove('lazy-image');
                        observer.unobserve(img);
                    }
                });
            },
            { threshold: 0.1 }
        );

        this.elements.resultsContainer.querySelectorAll('.lazy-image')
            .forEach(img => observer.observe(img));
    }
}

// Initialize the app
document.addEventListener('DOMContentLoaded', () => {
    const app = new App();
    app.searchArtists('');
});