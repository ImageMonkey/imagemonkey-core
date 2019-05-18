package main

import(
	"database/sql"
	_ "github.com/lib/pq"
	log "github.com/Sirupsen/logrus"
	"flag"
	datastructures "./datastructures"
	commons "./commons"
)


func getLabelId(tx *sql.Tx, label string, sublabel string) int64 {
	var labelId int64
	labelId = -1
	if sublabel == "" {
		err := tx.QueryRow(`SELECT id FROM label WHERE name = $1 and parent_id is null`, label).Scan(&labelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't get label id: ", err)
		}
	} else {
		err := tx.QueryRow(`SELECT l.id FROM label l 
							JOIN label pl ON pl.id = l.parent_id
							WHERE l.name = $1 and pl.name = $2`, sublabel, label).Scan(&labelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't get label id: ", label, sublabel)
		}
	}

	return labelId
}

func addAccessor(tx *sql.Tx, labelId int64, accessor string) error {
	var insertedId int64
	insertedId = -1
	err := tx.QueryRow(`INSERT INTO label_accessor(label_id, accessor) VALUES($1, $2)
                       				ON CONFLICT (label_id, accessor) DO NOTHING RETURNING id`, labelId, accessor).Scan(&insertedId)

	if insertedId != -1 {
		log.Info("Inserted label accessor ", accessor, " for label with id ", labelId)
	} else {
		err = nil //a negative id means, that the label already exists
	}

	return err
}

func addAccessors(tx *sql.Tx, label string, val datastructures.LabelMapEntry) {
	for _, accessor := range val.Accessors {

		labelId := getLabelId(tx, label, "")
		if labelId == -1 {
			tx.Rollback()
			log.Fatal("Populating accessors...couldn't get label id")
		}

		accessorStr := accessor
		if accessor == "." {
			accessorStr = label
		}

		err := addAccessor(tx, labelId, accessorStr)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't add accessor: ", err.Error())
		}


		if len(val.LabelMapEntries) != 0 {
			for sublabel, subval  := range val.LabelMapEntries {
				for _, subaccessor := range subval.Accessors {
					labelId := getLabelId(tx, label, sublabel)
					if labelId == -1 {
						tx.Rollback()
						log.Fatal("Populating accessors...couldn't get label id")
					}

					subaccessorStr := accessorStr + subaccessor + "='" + sublabel + "'" 
					if subaccessor == "." {
						subaccessorStr = accessorStr + sublabel
					}

					err := addAccessor(tx, labelId, subaccessorStr)
					if err != nil {
						tx.Rollback()
						log.Fatal("Couldn't add accessor for sublabel: ", err.Error())
					}
				}
			}
		}


		//quiz
		for _, quizEntry := range val.Quiz {
			for _, quizEntryAccessor := range quizEntry.Accessors {
				for _, answer := range quizEntry.Answers {

					sublabel := answer.Name
					if label == answer.Name {
						sublabel = ""
					}

					quizEntryAccessorStr := accessorStr + quizEntryAccessor + "='" + answer.Name + "'"
					
					labelId := getLabelId(tx, label, sublabel)
					if labelId == -1 {
						tx.Rollback()
						log.Fatal("Populating accessors...couldn't get label id")
					}

					err := addAccessor(tx, labelId, quizEntryAccessorStr)
					if err != nil {
						tx.Rollback()
						log.Fatal("Couldn't add accessor for quiz entry: ", err.Error())
					}

				}
			}
		}
	}
}

func addQuizQuestion(tx *sql.Tx, parentLabelUuid string, val datastructures.LabelMapEntry) {
	for _, quizEntry := range val.Quiz {
		_, err := tx.Exec(`INSERT INTO quiz_question(question, refines_label_id, recommended_control, 
														allow_unknown, allow_other, browse_by_example, multiselect, uuid)
							SELECT $1, id, $2, $3, $4, $5, $6, $8 FROM label WHERE uuid = $7
							ON CONFLICT(uuid) DO NOTHING`, 
							quizEntry.Question, quizEntry.ControlType, quizEntry.AllowUnknown, 
							quizEntry.AllowOther, quizEntry.BrowseByExample, quizEntry.Multiselect,
							parentLabelUuid, quizEntry.Uuid)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't add quiz question ", err.Error())
		}
	}
}

func addQuizAnswers(tx *sql.Tx, parentLabelUuid string, val datastructures.LabelMapEntry) {
	//quiz answers
	for _, quizEntry := range val.Quiz {
		for _, answer := range quizEntry.Answers {
			rows, err := tx.Query(`INSERT INTO label(name, parent_id, uuid, label_type)
								SELECT $1, id, $3, 'refinement' FROM label WHERE uuid = $2
								ON CONFLICT(uuid) DO NOTHING RETURNING id`, answer.Name, parentLabelUuid, answer.Uuid)
			if err != nil {
				tx.Rollback()
				log.Fatal("Couldn't add quiz answer label ", err.Error())
			}

			defer rows.Close()

			if rows.Next() {
				log.Info("Added quiz answer label ", answer.Name)
			}

			rows.Close()

			rows, err = tx.Query(`INSERT INTO quiz_answer(quiz_question_id, label_id)
									SELECT (SELECT q.id FROM quiz_question q WHERE q.uuid = $1),
									   	   (SELECT l.id FROM label l WHERE l.uuid = $2)
								   ON CONFLICT DO NOTHING RETURNING id`, quizEntry.Uuid, answer.Uuid)

			if err != nil {
				tx.Rollback()
				log.Fatal("Couldn't add quiz answer entry ", err.Error())
			}

			defer rows.Close()
				
			if rows.Next() {
				log.Info("Added quiz answer entry for label ", answer.Name)
			}

			rows.Close()
		}
	}
}

func addLabel(tx *sql.Tx, uuid string, label string) {
	rows, err := tx.Query("SELECT COUNT(id) FROM label WHERE uuid = $1", uuid)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	if !rows.Next() {
		tx.Rollback()
		log.Fatal(err)
	}

	numOfLabels := 0
	err = rows.Scan(&numOfLabels)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	rows.Close()

	if numOfLabels == 0 {
		log.Info("Adding label ", label)
		_,err := tx.Exec("INSERT INTO label(name, uuid, label_type) VALUES($1, $2, 'normal')", label, uuid)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	} else {
		log.Debug("Skipping label ", label, " as it already exists")
	}
}

func addSublabel(tx *sql.Tx, uuid string, label string, sublabel string) {
	rows, err := tx.Query("SELECT count(*) FROM label WHERE uuid = $1", uuid)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	if !rows.Next() {
		tx.Rollback()
		log.Fatal(err)
	}

	numOfLabels := 0
	err = rows.Scan(&numOfLabels)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	rows.Close()

	if numOfLabels == 0 {
		log.Info("Adding label ", sublabel, " (parent: ", label, " )")
		_,err := tx.Exec(`INSERT INTO label(name, parent_id, uuid, label_type)
			SELECT $1, l.id, $3, 'normal' FROM label l WHERE l.name = $2 AND l.parent_id is null`,
			sublabel, label, uuid)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	} else {
		log.Debug("Skipping label ", sublabel, " (parent: ", label, " ), as it already exists")
	}
}

func addLabelRefinements(tx *sql.Tx, labelMapRefinementEntries map[string]datastructures.LabelMapRefinementEntry) {
	for k, v := range labelMapRefinementEntries {
		if v.Uuid == "" {
			tx.Rollback()
			log.Fatal("refinement type uuid is empty!")
		}

		rows, err := tx.Query(`INSERT INTO label(name, parent_id, uuid, label_type) VALUES ($1, null, $2, 'refinement_category')
				                   ON CONFLICT DO NOTHING RETURNING id`, k, v.Uuid)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		defer rows.Close()

		if rows.Next() {
			log.Info("Inserted refinement type ", k, "(uuid: ", v.Uuid, ")")
		}

		rows.Close()

		for k1, v1 := range v.Values {
			if v1.Uuid == "" {
				tx.Rollback()
				log.Fatal("refinement label uuid is empty!")
			}

			//insert label if not exists
			rows, err = tx.Query(`INSERT INTO label(name, parent_id, uuid, label_type)
									SELECT $1, l.id, $2, 'refinement' FROM label l WHERE l.uuid = $3
				                   ON CONFLICT DO NOTHING RETURNING id`, k1, v1.Uuid, v.Uuid)

			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			defer rows.Close()

			if rows.Next() {
				log.Info("Inserted label ", k1, "(uuid: ", v1.Uuid, ")")
			}

			rows.Close()

			addLabelRefinementAccessors(tx, v1.Uuid, k1, v1.Accessors)
		}
	}
}

func addLabelRefinementAccessors(tx *sql.Tx, labelUuid string, label string, accessors []string) {
	for _, acc := range accessors {
		accessor := acc
		if accessor == "." {
			accessor = label
		}

		rows, err := tx.Query(`INSERT INTO label_accessor(label_id, accessor)
								SELECT l.id, $2 FROM label l WHERE uuid = $1
								ON CONFLICT DO NOTHING RETURNING id`, labelUuid, accessor)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		defer rows.Close()

		if rows.Next() {
			log.Info("Added label accessor ", accessor, " for label with uuid ", labelUuid)
		}

		rows.Close()
	}
}

func addMetaLabels(tx *sql.Tx, metaLabels datastructures.MetaLabelMap) {
	for k, v := range metaLabels.MetaLabelMapEntries {
		if v.Uuid == "" {
			tx.Rollback()
			log.Fatal("metalabels uuid is empty!")
		}

		rows, err := tx.Query(`INSERT INTO label(name, parent_id, uuid, label_type) VALUES ($1, null, $2, 'meta')
				                   ON CONFLICT DO NOTHING RETURNING id`, k, v.Uuid)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		defer rows.Close()

		if rows.Next() {
			log.Info("Inserted meta label type ", k, "(uuid: ", v.Uuid, ")")
		}

		rows.Close()

		addMetaLabelAccessors(tx, v.Uuid, k, v.Accessors)
	}
}


func addMetaLabelAccessors(tx *sql.Tx, labelUuid string, label string, accessors []string) {
	for _, acc := range accessors {
		accessor := acc
		if accessor == "." {
			accessor = label
		}

		rows, err := tx.Query(`INSERT INTO label_accessor(label_id, accessor)
								SELECT l.id, $2 FROM label l WHERE uuid = $1
								ON CONFLICT DO NOTHING RETURNING id`, labelUuid, accessor)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		defer rows.Close()

		if rows.Next() {
			log.Info("Added label accessor ", accessor, " for label with uuid ", labelUuid)
		}

		rows.Close()
	}
}

func addLabels(tx *sql.Tx, labelMap map[string]datastructures.LabelMapEntry) {
	for k := range labelMap {
		val := labelMap[k]

		addLabel(tx, val.Uuid, k)

		if len(val.LabelMapEntries) != 0 {
			for sublabel := range val.LabelMapEntries {
				addSublabel(tx, val.LabelMapEntries[sublabel].Uuid, k, sublabel)	
			}

		}

		addQuizQuestion(tx, val.Uuid, val)
		addQuizAnswers(tx, val.Uuid, val)
		addAccessors(tx, k, val)

	} 
}

func main(){
	dryRun := flag.Bool("dryrun", true, "dry run")
	debug := flag.Bool("debug", false, "debug")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *dryRun {
		log.Info("Populating labels (dry run)...")
	} else{
		log.Info("Populating labels...")
	}

	labelRepository := commons.NewLabelRepository()
	err := labelRepository.Load("../wordlists/en/labels.jsonnet")
	if err != nil {
		log.Fatal("Couldn't get label map: ", err.Error())
	}
    labelMap := labelRepository.GetMapping()
	

	labelMapRefinements, err := commons.GetLabelRefinementsMap("../wordlists/en/label-refinements.json")
	if err != nil {
		log.Fatal("Couldn't get label map refinements: ", err.Error())
	}

	metalabels := commons.NewMetaLabels("../wordlists/en/metalabels.jsonnet")
	err = metalabels.Load()
	if err != nil {
		log.Fatal("Couldn't get meta labels: ", err.Error())
	}


	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
    if err != nil {
    	log.Fatal("Couldn't start transaction: ", err.Error())
    }

	addLabels(tx, labelMap)
	addLabelRefinements(tx, labelMapRefinements)
	addMetaLabels(tx, metalabels.GetMapping())

	if ! *dryRun {
		err = tx.Commit()
		if err != nil {
			log.Fatal("Couldn't commit changes: ", err.Error())
		}
	} else {
		tx.Rollback()
		log.Info("Rolling back transaction...it was only a dry run.")
	}
}
