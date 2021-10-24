describe('Unified Mode', () => {
    beforeEach(() => {
        cy.clear_db_and_create_moderator_account();
        cy.donate_image('apple1.jpeg');
        cy.donate_image('apple2.jpeg');
        cy.login('moderator', 'moderator');
        cy.unlock_all_images();
    })

    it('Browse Image', () => {
        cy.query_images("image.unlabeled='true'", 2);
    });

    it('Open Image for Annotation', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
            cy.get('#loading-spinner').should('not.be.visible');
        });
    });

    it('Label Image', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
            cy.get('#loading-spinner').should('not.be.visible');
            //add label 'apple'
            cy.get('#add-labels-input').type('apple');
            cy.get('#add-labels-input').type('{enter}');
            //add label 'banana'
            cy.get('#add-labels-input').type('banana');
            cy.get('#add-labels-input').type('{enter}');
            cy.get('#annotation-label-list').find('table').find('td').should('have.length', 2);
        });
    });

    it('Remove Label again', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
            cy.get('#loading-spinner').should('not.be.visible');
            //add label 'apple'
            cy.get('#add-labels-input').type('apple');
            cy.get('#add-labels-input').type('{enter}');
            cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
            //remove label 'apple' again
            cy.get('#annotation-label-list').find('table').find('td').find('button').click();
            cy.get('#remove-label-confirmation-dialog').find('button').contains('Remove').click();
            cy.get('#annotation-label-list').find('table').find('td').should('have.length', 0);
        });
    });

    it('Annotate Image', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
            cy.get('#loading-spinner').should('not.be.visible');
            //add label 'apple'
            cy.get('#add-labels-input').type('apple');
            cy.get('#add-labels-input').type('{enter}');
            cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
            cy.draw_rectangle(0, 0, 200, 100);
        });
    });

    it('Annotate Image and Browse', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().click();
        cy.get('#loading-spinner').should('not.be.visible');
        //add label 'apple'
        cy.get('#add-labels-input').type('apple');
        cy.get('#add-labels-input').type('{enter}');
        cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
        cy.draw_rectangle(0, 0, 200, 100);
        cy.get('#annotation-navbar').find('button').contains('Save').click();

        cy.get('#annotation-image-grid').find('img').first().parent().should('have.class', 'grey-out');
    });

    it('Annotate Image, Discard and Browse', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().then((elem) => {
            elem.click();
            cy.get('#loading-spinner').should('not.be.visible');
            //add label 'apple'
            cy.get('#add-labels-input').type('apple');
            cy.get('#add-labels-input').type('{enter}');
            cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
            cy.draw_rectangle(0, 0, 200, 100);
            cy.get('#annotation-navbar').find('button').contains('Discard').click();

            cy.get('#annotation-image-grid').find('img').first().parent().should('not.have.class', 'grey-out');
            cy.query_images('apple', 0);
        });
    });

    it('Label Image and Browse', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().click();
        cy.get('#loading-spinner').should('not.be.visible');
        //add label 'apple'
        cy.get('#add-labels-input').type('apple');
        cy.get('#add-labels-input').type('{enter}');
        cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
        cy.get('#annotation-navbar').find('button').contains('Save').click();

        cy.get('#annotation-image-grid').find('img').first().parent().should('have.class', 'grey-out');
        cy.query_images('apple', 1);
    });

    it('Annotate Image, Browse and Query annotated images', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().click();
        cy.get('#loading-spinner').should('not.be.visible');
        //add label 'apple'
        cy.get('#add-labels-input').type('apple');
        cy.get('#add-labels-input').type('{enter}');
        cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
        cy.draw_rectangle(0, 0, 200, 100);
        cy.get('#annotation-navbar').find('button').contains('Save').click();

        cy.get('#annotation-image-grid').find('img').first().parent().should('have.class', 'grey-out');

        cy.query_annotated_images('apple', 1);
    });

    it('Duplicate Label', () => {
        cy.query_images("image.unlabeled='true'", 2);
        cy.get('#annotation-image-grid').find('img').first().click();
        cy.get('#loading-spinner').should('not.be.visible');
        //add label 'apple'
        cy.get('#add-labels-input').type('apple');
        cy.get('#add-labels-input').type('{enter}');
        cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
        cy.get('#add-labels-input').type('apple');
        cy.get('#add-labels-input').type('{enter}');
        cy.get('#simple-error-popup').contains('Label apple already exists');
		cy.get('#annotation-label-list').find('table').find('td').should('have.length', 1);
    });
})
