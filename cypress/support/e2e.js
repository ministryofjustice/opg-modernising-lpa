// ***********************************************************
// This example support/e2e.js is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands'

// Alternatively you can use CommonJS syntax:
// require('./commands')
import 'cypress-axe'

export const AddressFormAssertions = {
    assertCanAddAddressManually(manualAddressLinkText) {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('a', manualAddressLinkText).click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').type('Flat 2');
        cy.get('#f-address-line-2').type('123 Fake Street');
        cy.get('#f-address-line-3').type('Pretendingham');
        cy.get('#f-address-town').type('Someville');
        cy.get('#f-address-postcode').type('NG1');

        cy.contains('button', 'Continue').click();
    },

    assertCanAddAddressFromSelect() {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').should('have.value', '2 RICHMOND PLACE');
        cy.get('#f-address-line-2').should('have.value', '');
        cy.get('#f-address-line-3').should('have.value', '');
        cy.get('#f-address-town').should('have.value', 'BIRMINGHAM');
        cy.get('#f-address-postcode').should('have.value', 'B14 7ED');
        cy.contains('button', 'Continue').click();
    },

    assertErrorsWhenPostcodeEmpty() {
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter a postcode');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter a postcode');
    },

    assertErrorsWhenUnselected() {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select an address from the list');
        });

        cy.contains('[for=f-select-address] + .govuk-error-message', 'Select an address from the list');
    },

    assertErrorsWhenManualIncorrect(manualAddressLinkText) {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();
        cy.contains('a', manualAddressLinkText).click();
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter address line 1');
            cy.contains('Enter town or city');
            cy.contains('Enter a postcode');
        });

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Enter address line 1');
        cy.contains('[for=f-address-town] + .govuk-error-message', 'Enter town or city');
        cy.contains('[for=f-address-postcode] + .govuk-error-message', 'Enter a postcode');

        cy.get('#f-address-line-1').invoke('val', 'a'.repeat(51));
        cy.get('#f-address-line-2').invoke('val', 'b'.repeat(51));
        cy.get('#f-address-line-3').invoke('val', 'c'.repeat(51));
        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Address line 1 must be 50 characters or less');
        cy.contains('[for=f-address-line-2] + .govuk-error-message', 'Address line 2 must be 50 characters or less');
        cy.contains('[for=f-address-line-3] + .govuk-error-message', 'Address line 3 must be 50 characters or less');
    },

    assertErrorsWhenInvalidPostcode() {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select an address from the list');
        });

        cy.contains('[for=f-select-address] + .govuk-error-message', 'Select an address from the list');
    }
}
