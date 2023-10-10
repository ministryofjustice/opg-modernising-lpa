import { AddressFormAssertions } from "../../support/e2e";

describe('People to notify address', () => {
    describe('Entering a new address', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/choose-people-to-notify-address?id=without-address&progress=peopleToNotifyAboutYourLpa&peopleToNotify=without-address');
            cy.contains('label', 'Enter a new address').click();
            cy.contains('button', 'Continue').click();
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
        cy.visit('/fixtures?redirect=/choose-people-to-notify-address?id=without-address&progress=peopleToNotifyAboutYourLpa&peopleToNotify=without-address');
        cy.contains('label', 'Use an address you’ve already entered').click();
        cy.contains('button', 'Continue').click();

        cy.contains('label', '5 RICHMOND PLACE').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.get('#address-1').should('contain', '5 RICHMOND PLACE');
    });
});
