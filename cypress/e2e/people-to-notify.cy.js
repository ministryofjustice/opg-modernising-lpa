describe('People to notify', () => {
    let person1
    let person2

    before(() => {
        cy.fixture('peopleToNotify.json').then(p => {
            person1 = p.person1
            person2 = p.person2
        })
    })

    it('can add people to notify', () => {
        cy.visit('/testing-start?redirect=/do-you-want-to-notify-people');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('input[name="want-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        addPersonToNotify(person1)

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('You have added 1 person to notify');

        cy.get('#name-1').contains(`${person1.firstNames} ${person1.lastName}`);
        cy.get('#address-1').contains(person1.address.line1);
        cy.get('#address-1').contains(person1.address.town);
        cy.get('#address-1').contains(person1.address.postcode);

        cy.get('input[name="add-person-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        addPersonToNotify(person2)

        cy.contains('You have added 2 people to notify');

        cy.get('#name-2').contains(`${person2.firstNames} ${person2.lastName}`);
        cy.get('#address-2').contains(person2.address.line1);
        cy.get('#address-2').contains(person2.address.town);
        cy.get('#address-2').contains(person2.address.postcode);

        cy.get('input[name="add-person-to-notify"]').check('no')
        cy.contains('button', 'Continue').click();

        cy.visit('/task-list')

        cy.contains('a', 'People to notify').parent().parent().contains('Completed (2)')
    });

    it('can modify a person to notifys details', () => {
        cy.visit('/testing-start?redirect=/do-you-want-to-notify-people');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('input[name="want-to-notify"]').check('yes')
        cy.contains('button', 'Continue').click();

        addPersonToNotify(person1)

        cy.contains(`${person1.firstNames} ${person1.lastName}`).parent().contains('a', 'Change').click();

        cy.url().should('contain', '/choose-people-to-notify');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-first-names').clear().type('Changed')
        cy.get('#f-last-name').clear().type('Altered')
        cy.get('#f-email').clear().type('different@example.org')

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.get('#name-1').contains('Changed Altered')
        cy.get('#email-1').contains('different@example.org')

        cy.contains(person1.address.line1).parent().contains('a', 'Change').click();

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
            cy.visit('/testing-start?redirect=/do-you-want-to-notify-people');

            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.get('input[name="want-to-notify"]').check('yes')
            cy.contains('button', 'Continue').click();

            addPersonToNotify(person1)

            cy.get('input[name="add-person-to-notify"]').check('yes')
            cy.contains('button', 'Continue').click();

            addPersonToNotify(person2)

            cy.get('#remove-person-to-notify-2').contains('Remove person to notify 2').click();

            cy.url().should('contain', '/remove-person-to-notify');

            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.get('input[name="remove-person-to-notify"]').check('yes')
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/choose-people-to-notify-summary');

            cy.get('#remove-person-to-notify-1').contains('Remove person to notify 1').click();

            cy.url().should('contain', '/remove-person-to-notify');

            cy.get('input[name="remove-person-to-notify"]').check('yes')
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/do-you-want-to-notify-people');
    });
});

function addPersonToNotify(p) {
    cy.url().should('contain', '/choose-people-to-notify');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-first-names').type(p.firstNames)
    cy.get('#f-last-name').type(p.lastName)
    cy.get('#f-email').type(p.email)

    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-address');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-lookup-postcode').type(p.address.postcode)
    cy.contains('button', 'Find address').click();

    cy.url().should('contain', '/choose-people-to-notify-address');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-select-address').select(`${p.address.line1}, ${p.address.town}, ${p.address.postcode}`);
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-address');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-address-line-1').should('have.value', p.address.line1);
    cy.get('#f-address-line-2').should('have.value', p.address.line2);
    cy.get('#f-address-line-3').should('have.value', p.address.line3);
    cy.get('#f-address-town').should('have.value', p.address.town);
    cy.get('#f-address-postcode').should('have.value', p.address.postcode);

    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-summary');
}
