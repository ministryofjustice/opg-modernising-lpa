import {AddressFormAssertions, TestEmail, TestEmail2} from "../support/e2e";

describe('People to notify', () => {
    let person1

    before(() => {
        cy.fixture('peopleToNotify.json').then(p => {
            person1 = p.person1
        })
    })

    it('can add people to notify', () => {
        cy.visit('/testing-start?redirect=/do-you-want-to-notify-people&withDonorDetails=1&withAttorney=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('input[name="want-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        addPersonToNotify(person1, true)

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('People to notify about your LPA');

        cy.get('#name-1').contains(`${person1.firstNames} ${person1.lastName}`);
        cy.get('#address-1').contains(person1.address.line1);
        cy.get('#address-1').contains(person1.address.town);
        cy.get('#address-1').contains(person1.address.postcode);

        cy.get('input[name="add-person-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/task-list')

        cy.contains('a', 'People to notify').parent().parent().contains('Completed (1)')
    });

    it('can modify a person to notifys details', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify-summary&withDonorDetails=1&withAttorney=1&withPeopleToNotify=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Joanna Smith').parent().contains('a', 'Change').click();

        cy.url().should('contain', '/choose-people-to-notify');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-first-names').clear().type('Changed')
        cy.get('#f-last-name').clear().type('Altered')
        cy.get('#f-email').clear().type(TestEmail2)

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.get('#name-1').contains('Changed Altered')
        cy.get('#email-1').contains(TestEmail2)

        cy.contains('4 RICHMOND PLACE').parent().contains('a', 'Change').click();

        cy.url().should('contain', '/choose-people-to-notify-address');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').clear().type('1 New Road');
        cy.get('#f-address-line-2').clear().type('Changeville');
        cy.get('#f-address-line-3').clear().type('Newington');
        cy.get('#f-address-town').clear().type('Newshire');
        cy.get('#f-address-postcode').clear().type('A12 3BC');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.get('#address-1').contains('1 New Road');
        cy.get('#address-1').contains('Changeville');
        cy.get('#address-1').contains('Newington');
        cy.get('#address-1').contains('Newshire');
        cy.get('#address-1').contains('A12 3BC');
    });

    it('can remove a person to notify', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify-summary&withDonorDetails=1&withAttorney=1&withPeopleToNotify=2');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#remove-person-to-notify-2').contains(`Remove Jonathan Smith`).click();

        cy.url().should('contain', '/remove-person-to-notify');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('input[name="remove-person-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.get('#remove-person-to-notify-1').contains(`Remove Joanna Smith`).click();

        cy.url().should('contain', '/remove-person-to-notify');

        cy.get('input[name="remove-person-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-to-notify-people');
    });

    it('hides action links when LPA has been signed', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify-summary&completeLpa=1&withPeopleToNotify=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Joanna Smith').parent().contains('a', 'Change').should('not.exist');
    });

    it('limits people to notify to 5', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify-summary&completeLpa=1&withPeopleToNotify=5');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Do you want to add another person to notify?').should('not.exist');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/check-your-lpa');
    })

    it('errors when unselected', () => {
        cy.visit('/testing-start?redirect=/do-you-want-to-notify-people&withDonorDetails=1&withAttorney=1');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes to notify someone about your LPA');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to notify someone about your LPA');
    });

    it('errors when people to notify details empty', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter email address');
        });

        cy.contains('[for=f-first-names] + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('[for=f-email] + .govuk-error-message', 'Enter email address');
    });

    it('errors when people to notify details invalid', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify');

        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));
        cy.get('#f-email').type('not-an-email');
        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-first-names] + .govuk-error-message', 'First names must be 53 characters or less');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
        cy.contains('[for=f-email] + .govuk-error-message', 'Email address must be in the correct format, like name@example.com');
    });

    it('errors when another not selected', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify-summary&withDonorDetails=1&withAttorney=1&withPeopleToNotify=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes to add another person to notify');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to add another person to notify');
    });

    it('warns when name shared with other actor', () => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify&withDonorDetails=1');

        cy.get('#f-first-names').type('Jose');
        cy.get('#f-last-name').type('Smith');
        cy.get('#f-email').type(TestEmail);
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-people-to-notify');

        cy.contains('The donor’s name is also Jose Smith.');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-people-to-notify-address');
    });
});

function addPersonToNotify(p, manualAddress) {
    cy.url().should('contain', '/choose-people-to-notify');

    cy.injectAxe();
    cy.checkA11y(null, {rules: {region: {enabled: false}}});

    cy.get('#f-first-names').type(p.firstNames)
    cy.get('#f-last-name').type(p.lastName)
    cy.get('#f-email').type(p.email)

    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-address');
    cy.injectAxe();
    cy.checkA11y(null, {rules: {region: {enabled: false}}});

    cy.get('#f-lookup-postcode').type(p.address.postcode)
    cy.contains('button', 'Find address').click();

    cy.url().should('contain', '/choose-people-to-notify-address');
    cy.injectAxe();
    cy.checkA11y(null, {rules: {region: {enabled: false}}});

    if (manualAddress) {
        cy.contains('a', "I can’t find their address in the list").click();

        cy.url().should('contain', '/choose-people-to-notify-address');
        cy.injectAxe();
        cy.checkA11y(null, {rules: {region: {enabled: false}}});

        cy.get('#f-address-line-1').type(p.address.line1);
        cy.get('#f-address-town').type(p.address.town);
        cy.get('#f-address-postcode').type(p.address.postcode);
    } else {
        cy.injectAxe();
        cy.checkA11y(null, {rules: {region: {enabled: false}}});

        cy.get('#f-select-address').select(`${p.address.line1}, ${p.address.town}, ${p.address.postcode}`);
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-address');

        cy.injectAxe();
        cy.checkA11y(null, {rules: {region: {enabled: false}}});

        cy.get('#f-address-line-1').should('have.value', p.address.line1);
        cy.get('#f-address-line-2').should('have.value', p.address.line2);
        cy.get('#f-address-line-3').should('have.value', p.address.line3);
        cy.get('#f-address-town').should('have.value', p.address.town);
        cy.get('#f-address-postcode').should('have.value', p.address.postcode);
    }

    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-summary');
}
