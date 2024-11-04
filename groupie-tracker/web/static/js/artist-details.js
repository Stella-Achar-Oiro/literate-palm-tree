// web/static/js/artist-details.js

// Constants and Globals
const MAPBOX_TOKEN = 'pk.eyJ1Ijoic3RlbGxhYWNoYXJvaXJvIiwiYSI6ImNtMWhmZHNlODBlc3cybHF5OWh1MDI2dzMifQ.wk3v-v7IuiSiPwyq13qdHw';
let map;
let activePopup = null;
let favorites = JSON.parse(localStorage.getItem('favorites')) || [];

// DOM Elements
const elements = {
    artistDetails: document.getElementById('artist-details'),
    map: document.getElementById('map'),
    loading: document.getElementById('loading'),
    errorMessage: document.getElementById('error-message')
};

// Initialize Mapbox
mapboxgl.accessToken = MAPBOX_TOKEN;

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

function getArtistId() {
    const pathParts = window.location.pathname.split('/');
    return pathParts[pathParts.length - 1];
}

// Favorites Management
function toggleFavorite(artistId) {
    const index = favorites.indexOf(artistId);
    if (index === -1) {
        favorites.push(artistId);
    } else {
        favorites.splice(index, 1);
    }
    localStorage.setItem('favorites', JSON.stringify(favorites));
    updateFavoriteButton(artistId);
}

function updateFavoriteButton(artistId) {
    const button = document.getElementById(`favorite-${artistId}`);
    if (button) {
        button.innerHTML = favorites.includes(artistId)
            ? '<i class="fas fa-star"></i> Remove from Favorites'
            : '<i class="far fa-star"></i> Add to Favorites';
    }
}

// Sharing Function
async function shareArtist(artist) {
    if (navigator.share) {
        try {
            await navigator.share({
                title: artist.name,
                text: `Check out ${artist.name} on Groupie Tracker!`,
                url: window.location.href
            });
        } catch (error) {
            if (error.name !== 'AbortError') {
                console.error('Error sharing:', error);
                showError('Failed to share artist information');
            }
        }
    } else {
        // Fallback for browsers that don't support Web Share API
        const textArea = document.createElement('textarea');
        textArea.value = window.location.href;
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            alert('Link copied to clipboard!');
        } catch (err) {
            console.error('Failed to copy:', err);
            showError('Failed to copy link to clipboard');
        }
        document.body.removeChild(textArea);
    }
}

// Map Functions
function initializeMap() {
    if (map) {
        map.remove();
    }

    map = new mapboxgl.Map({
        container: 'map',
        style: 'mapbox://styles/mapbox/dark-v10',
        center: [0, 20],
        zoom: 1.5
    });

    map.addControl(new mapboxgl.NavigationControl());
}

function displayMap(locations) {
    initializeMap();
    const bounds = new mapboxgl.LngLatBounds();

    locations.forEach(location => {
        if (!location.lon || !location.lat) {
            console.error('Invalid coordinates for location:', location);
            return;
        }

        // Create custom marker element
        const el = document.createElement('div');
        el.className = 'custom-marker';
        el.innerHTML = '<i class="fas fa-map-marker-alt"></i>';
        el.style.color = '#FF0000';
        el.style.fontSize = '24px';
        el.style.cursor = 'pointer';

        // Add marker to map
        const marker = new mapboxgl.Marker({ element: el })
            .setLngLat([location.lon, location.lat])
            .addTo(map);

        // Add popup
        const popup = createPopup(location);
        
        // Add click event
        el.addEventListener('click', () => {
            if (activePopup) {
                activePopup.remove();
            }
            popup.setLngLat([location.lon, location.lat])
                .addTo(map);
            activePopup = popup;
        });

        // Add hover effects
        el.addEventListener('mouseenter', () => {
            el.style.transform = 'scale(1.2)';
        });

        el.addEventListener('mouseleave', () => {
            el.style.transform = 'scale(1)';
        });

        bounds.extend([location.lon, location.lat]);
    });

    if (!bounds.isEmpty()) {
        map.fitBounds(bounds, {
            padding: 50,
            duration: 1000
        });
    }

    // Close popup when clicking on map
    map.on('click', () => {
        if (activePopup) {
            activePopup.remove();
            activePopup = null;
        }
    });
}

function createPopup(location) {
    return new mapboxgl.Popup({
        closeButton: true,
        closeOnClick: false,
        className: 'custom-popup'
    }).setHTML(`
        <div class="popup-content">
            <h3>${location.address}</h3>
            <div class="popup-body">
                ${location.dates ? `
                    <h4>Concert Dates:</h4>
                    <ul>
                        ${location.dates.map(date => `<li>${date}</li>`).join('')}
                    </ul>
                ` : ''}
            </div>
        </div>
    `);
}

// Display Functions
function displayArtistDetails(details) {
    elements.artistDetails.innerHTML = `
        <div class="artist-header">
            <img src="${details.artist.image}" alt="${details.artist.name}" class="artist-image">
            <div class="artist-info">
                <h2>${details.artist.name}</h2>
                <p><i class="fas fa-users"></i> Members: ${details.artist.members.join(', ')}</p>
                <p><i class="fas fa-calendar-alt"></i> Creation Date: ${details.artist.creationDate}</p>
                <p><i class="fas fa-compact-disc"></i> First Album: ${details.artist.firstAlbum}</p>
            </div>
        </div>

        <div class="artist-content">
            <div class="locations-section">
                <h3><i class="fas fa-map-marker-alt"></i> Concert Locations</h3>
                <ul class="locations-list">
                    ${details.locations.map(loc => `
                        <li>${loc.address}</li>
                    `).join('')}
                </ul>
            </div>

            <div class="dates-section">
                <h3><i class="fas fa-calendar-check"></i> Concert Dates</h3>
                <ul class="dates-list">
                    ${details.dates.map(date => `
                        <li>${date}</li>
                    `).join('')}
                </ul>
            </div>

            <div class="relations-section">
                <h3><i class="fas fa-link"></i> Location-Date Relations</h3>
                <ul class="relations-list">
                    ${Object.entries(details.relations).map(([loc, dates]) => `
                        <li>
                            <strong>${loc}:</strong>
                            <span>${dates.join(', ')}</span>
                        </li>
                    `).join('')}
                </ul>
            </div>
        </div>

        <div class="action-buttons">
            <button id="favorite-${details.artist.id}" 
                    onclick="toggleFavorite(${details.artist.id})" 
                    class="favorite-button">
                ${favorites.includes(details.artist.id) 
                    ? '<i class="fas fa-star"></i> Remove from Favorites' 
                    : '<i class="far fa-star"></i> Add to Favorites'}
            </button>
            <button onclick='shareArtist(${JSON.stringify(details.artist)})' 
                    class="share-button">
                <i class="fas fa-share-alt"></i> Share
            </button>
        </div>
    `;

    displayMap(details.locations);
}

// Initial Load
window.addEventListener('load', async () => {
    const artistId = getArtistId();
    if (artistId) {
        showLoading();
        try {
            const response = await fetch(`/api/artist/${artistId}`);
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            const data = await response.json();
            displayArtistDetails(data);
        } catch (error) {
            console.error('Error:', error);
            showError('An error occurred while fetching artist details. Please try again later.');
        } finally {
            hideLoading();
        }
    }
});