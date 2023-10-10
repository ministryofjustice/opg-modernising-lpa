import { AddressFormAssertions } from "../../support/e2e";

describe('Enter replacement trust corporation address', () => {
    describe('Entering a new address', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/enter-replacement-trust-corporation-address&progress=chooseYourReplacementAttorneys&replacementAttorneys=trust-corporation-without-address');
            cy.contains('label', 'Enter a new address').click();
            cy.contains('button', 'Continue').click();
        });

        it('address can be looked up', () => {
            AddressFormAssertions.assertCanAddAddressFromSelect()
            cy.url().should('contain', '/choose-replacement-attorneys-summary');
        });

        it('address can be entered manually if not found', () => {
            AddressFormAssertions.assertCanAddAddressManually('I can’t find their address in the list')
            cy.url().should('contain', '/choose-replacement-attorneys-summary');
        });

        it('address can be entered manually on invalid postcode', () => {
            AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)
            cy.url().should('contain', '/choose-replacement-attorneys-summary');
        });

        it('errors when empty postcode', () => {
            AddressFormAssertions.assertErrorsWhenPostcodeEmpty()
        });

        it('errors when invalid postcode', () => {
            AddressFormAssertions.assertErrorsWhenInvalidPostcode()
        });

        it('errors when valid postcode and no addresses', () => {
            AddressFormAssertions.assertErrorsWhenValidPostcodeFormatButNoAddressesFound()
        });

        it('errors when unselected', () => {
            AddressFormAssertions.assertErrorsWhenUnselected()
        });

        it('errors when manual incorrect', () => {
            AddressFormAssertions.assertErrorsWhenManualIncorrect('I can’t find their address in the list')
        });
    });

    it('address can be copied from another actor', () => {
        cy.visit('/fixtures?redirect=/enter-replacement-trust-corporation-address&progress=chooseYourReplacementAttorneys&replacementAttorneys=trust-corporation-without-address');
        cy.contains('label', 'Use an address you’ve already entered').click();
        cy.contains('button', 'Continue').click();

        cy.contains('label', '2 RICHMOND PLACE').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-summary');

        cy.get('#trust-corporation-address').should('contain', '2 RICHMOND PLACE');
    });
});
