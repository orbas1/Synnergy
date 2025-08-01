import { renderListings, initListingForm } from './components/listings.js';
import { renderDeals, initDealForm } from './components/deals.js';

async function init() {
  await renderListings();
  await renderDeals();
  initListingForm();
  initDealForm();
}

document.addEventListener('DOMContentLoaded', init);
