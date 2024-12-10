import './commands'
import 'cypress-axe'
import "cypress-real-events"

export const
    TestEmail = 'simulate-delivered@notifications.service.gov.uk',
    TestEmail2 = 'simulate-delivered-2@notifications.service.gov.uk',
    TestMobile = '07700900000',
    TestMobile2 = '07700900111';

export function randomShareCode() {
    const characters = 'abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789'
    let result = [];

    for (let i = 0; i < 12; i++) {
        result.push(characters.charAt(Math.floor(Math.random() * characters.length)));
    }

    return result.join('');
}

export const AddressFormAssertions = {
    assertCanAddAddressManually(manualAddressLinkText, withInvalidPostcode = false) {
        cy.checkA11yApp();

        if (withInvalidPostcode) {
            cy.get('#f-lookup-postcode').type('INVALID');
        } else {
            cy.get('#f-lookup-postcode').type('NG1');
        }

        cy.contains('button', 'Find address').click();

        cy.checkA11yApp();

        cy.contains('a', manualAddressLinkText).click();

        cy.checkA11yApp();

        cy.get('#f-address-line-1').type('Flat 2');
        cy.get('#f-address-line-2').type('123 Fake Street');
        cy.get('#f-address-line-3').type('Pretendingham');
        cy.get('#f-address-town').type('Someville');
        cy.get('#f-address-postcode').type('NG1');

        cy.contains('button', 'Save and continue').click();
    },

    assertCanAddAddressFromSelect() {
        cy.checkA11yApp();

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.checkA11yApp();

        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();

        cy.get('#f-address-line-1').should('have.value', '2 RICHMOND PLACE');
        cy.get('#f-address-line-2').should('have.value', '');
        cy.get('#f-address-line-3').should('have.value', '');
        cy.get('#f-address-town').should('have.value', 'BIRMINGHAM');
        cy.get('#f-address-postcode').should('have.value', 'B14 7ED');
        cy.contains('button', 'Save and continue').click();
    },

    assertErrorsWhenPostcodeEmpty() {
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter a postcode');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter a postcode');
    },

    assertErrorsWhenYourPostcodeEmpty() {
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your postcode');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter your postcode');
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

    assertErrorsWhenYourAddressUnselected() {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select your address from the list');
        });

        cy.contains('[for=f-select-address] + .govuk-error-message', 'Select your address from the list');
    },

    assertErrorsWhenManualIncorrect(manualAddressLinkText) {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();
        cy.contains('a', manualAddressLinkText).click();
        cy.contains('button', 'Save and continue').click();

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
        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Address line 1 must be 50 characters or less');
        cy.contains('[for=f-address-line-2] + .govuk-error-message', 'Address line 2 must be 50 characters or less');
        cy.contains('[for=f-address-line-3] + .govuk-error-message', 'Address line 3 must be 50 characters or less');
    },

    assertErrorsWhenYourManualIncorrect(manualAddressLinkText) {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();
        cy.contains('a', manualAddressLinkText).click();
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter line 1 of your address');
            cy.contains('Enter your town or city');
            cy.contains('Enter your postcode');
        });

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Enter line 1 of your address');
        cy.contains('[for=f-address-town] + .govuk-error-message', 'Enter your town or city');
        cy.contains('[for=f-address-postcode] + .govuk-error-message', 'Enter your postcode');

        cy.get('#f-address-line-1').invoke('val', 'a'.repeat(51));
        cy.get('#f-address-line-2').invoke('val', 'b'.repeat(51));
        cy.get('#f-address-line-3').invoke('val', 'c'.repeat(51));
        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Line 1 of your address must be 50 characters or less');
        cy.contains('[for=f-address-line-2] + .govuk-error-message', 'Line 2 of your address must be 50 characters or less');
        cy.contains('[for=f-address-line-3] + .govuk-error-message', 'Line 3 of your address must be 50 characters or less');
    },

    assertErrorsWhenInvalidPostcode() {
        cy.get('#f-lookup-postcode').type('INVALID');
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter a valid postcode');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter a valid postcode');
    },

    assertErrorsWhenValidPostcodeFormatButNoAddressesFound() {
        const validFormatPostcodeWithNoAddresses = 'NE234EE'

        cy.get('#f-lookup-postcode').type(validFormatPostcodeWithNoAddresses);
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('We could not find any addresses for that postcode. Check the postcode is correct, or enter the address manually.');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'We could not find any addresses for that postcode. Check the postcode is correct, or enter the address manually.');
    },

    assertErrorsWhenYourValidPostcodeFormatButNoAddressesFound() {
        const validFormatPostcodeWithNoAddresses = 'NE234EE'

        cy.get('#f-lookup-postcode').type(validFormatPostcodeWithNoAddresses);
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('We could not find any addresses for that postcode. Check your postcode is correct, or enter your address manually.');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'We could not find any addresses for that postcode. Check your postcode is correct, or enter your address manually.');
    }
}
