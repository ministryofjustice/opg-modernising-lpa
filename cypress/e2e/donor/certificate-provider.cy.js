import {
    AddressFormAssertions, TestEmail, TestMobile
} from "../../support/e2e";

describe('Certificate provider task', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys');
    });

    it('can be a professional', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.checkA11yApp();

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/choose-your-certificate-provider');
        cy.checkA11yApp();

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type(TestMobile);
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Professionally').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'By email').click();
        cy.get('#f-email').type(TestEmail, { force: true });
        cy.contains('button', 'Save and continue').click()

        cy.url().should('contain', '/certificate-provider-address');

        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.url().should('contain', '/task-list');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });

    it('can be a lay person', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.checkA11yApp();

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/choose-your-certificate-provider');
        cy.checkA11yApp();

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type(TestMobile);
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Personally').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/how-long-have-you-known-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How long have you known John Doe?');
        cy.contains('label', '2 years or more').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });
        cy.contains('label', 'Using paper forms').click();

        cy.contains('button', 'Save and continue').click()

        cy.url().should('contain', '/certificate-provider-address');
        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();

        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.url().should('contain', '/task-list');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });

    it('requires a new certificate provider when known for less than 2 years', () => {
        cy.visitLpa('/how-long-have-you-known-certificate-provider');

        cy.contains('label', 'Less than 2 years').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-new-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-your-certificate-provider');

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');

        cy.get('#f-first-names').should('have.value', '');
        cy.get('#f-last-name').should('have.value', '');
        cy.get('#f-mobile').should('have.value', '');
    });

    it('errors when details empty', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter your certificate provider’s UK mobile number');
        });

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('[for=f-mobile] + p + div + .govuk-error-message', 'Enter your certificate provider’s UK mobile number');
    });

    it('errors when invalid mobile number', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.get('#f-mobile').type('not-a-number');
        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-mobile] + p + div + .govuk-error-message', 'Enter a mobile number in the correct format');
    });

    it('errors when invalid non uk mobile number', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.get('#f-has-non-uk-mobile').check({ force: true });
        cy.get('#f-non-uk-mobile').type('not-a-number', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-non-uk-mobile] + div + .govuk-error-message', 'Enter a mobile number in the correct format');
    });

    it('errors when how they prefer to carry out their role unselected', () => {
        cy.visitLpa('/how-would-certificate-provider-prefer-to-carry-out-their-role');

        cy.contains('button', 'Save and continue').click()

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how your certificate provider would prefer to carry out their role');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how your certificate provider would prefer to carry out their role');
    });

    it('errors when how they prefer to carry out their role email invalid', () => {
        cy.visitLpa('/how-would-certificate-provider-prefer-to-carry-out-their-role');

        cy.contains('label', 'By email').click();
        cy.contains('button', 'Save and continue').click()
        cy.contains('[for=f-email] + .govuk-error-message', 'Enter certificate provider’s email address');

        cy.get('#f-email').type('not-an-email', { force: true });
        cy.contains('button', 'Save and continue').click()
        cy.contains('[for=f-email] + .govuk-error-message', 'Certificate provider’s email address must be in the correct format, like name@example.com');
    });

    it('errors when empty postcode', () => {
        cy.visitLpa('/certificate-provider-address');
        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertErrorsWhenPostcodeEmpty()
    });

    it('errors when unselected', () => {
        cy.visitLpa('/certificate-provider-address');
        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertErrorsWhenUnselected()
    });

    it('errors when manual incorrect', () => {
        cy.visitLpa('/certificate-provider-address');
        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertErrorsWhenManualIncorrect('I can’t find their address in the list')
    });

    it('errors when invalid postcode', () => {
        cy.visitLpa('/certificate-provider-address');
        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertErrorsWhenInvalidPostcode()
    });

    it('errors when how you know not selected', () => {
        cy.visitLpa('/how-do-you-know-your-certificate-provider');

        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how you know your certificate provider');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how you know your certificate provider');
    });

    it('errors how long you have known them not selected', () => {
        cy.visitLpa('/how-long-have-you-known-certificate-provider');

        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how long you have known your certificate provider');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how long you have known your certificate provider');
    });

    it('warns when name shared with other actor', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.contains('button', 'Save and continue').click();

        cy.get('#f-first-names').type('Jessie');
        cy.get('#f-last-name').type('Jones');
        cy.get('#f-mobile').type(TestMobile);
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/certificate-provider-details');

        cy.contains('There is also an attorney called Jessie Jones.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
    });
});
