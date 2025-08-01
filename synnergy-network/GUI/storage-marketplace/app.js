import { renderListings, initListingForm } from './components/listings.js';
import { renderDeals, initDealForm } from './components/deals.js';
import { initStorageSection } from './components/storage.js';

async function init() {
  await renderListings();
  await renderDeals();
  initListingForm();
  initDealForm();
  initStorageSection();
}

document.addEventListener('DOMContentLoaded', init);
