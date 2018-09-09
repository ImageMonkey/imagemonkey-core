package datastructures

import (
    "encoding/json"
)

type ImageAnnotationCoverage struct {
    Image struct {
        Id string `json:"id"`
        Width float32 `json:"width"`
        Height float32 `json:"height"`
    } `json:"image"`
    Coverage float32 `json:"coverage"`
}

type AnnotationStroke struct {
    Width float32 `json:"width"`
    Color string `json:"color"`
}

type RectangleAnnotation struct {
    //Id int64 `json:"id"`
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Width float32 `json:"width"`
    Height float32 `json:"height"`
    Angle float32 `json:"angle"`
    Type string `json:"type"`
    Stroke AnnotationStroke `json:"stroke"`
}

type EllipsisAnnotation struct {
    //Id int64 `json:"id"`
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Rx float32 `json:"rx"`
    Ry float32 `json:"ry"`
    Angle float32 `json:"angle"`
    Type string `json:"type"`
    Stroke AnnotationStroke `json:"stroke"`
}


type PolygonPoint struct {
    X float32 `json:"x"`
    Y float32 `json:"y"`
}

type PolygonAnnotation struct {
    //Id int64 `json:"id"`
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Points []PolygonPoint `json:"points"`
    Angle float32 `json:"angle"`
    Type string `json:"type"`
    Stroke AnnotationStroke `json:"stroke"`
}

type Annotations struct {
    Annotations []json.RawMessage `json:"annotations"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
}

type Image struct {
    Id string `json:"uuid"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
    Provider string `json:"provider"`
    Probability float32 `json:"probability"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
    Unlocked bool `json:"unlocked,omitempty"`
    Width int32 `json:"width,omitempty"`
    Height int32 `json:"height,omitempty"`
    Annotations []json.RawMessage `json:"annotations"`
    AllLabels []LabelMeEntry `json:"all_labels"`
}

type ValidationImage struct {
    Id string `json:"uuid"`
    Provider string `json:"provider"`
    Unlocked bool `json:"unlocked"`

    Label string `json:"label"`
    Sublabel string `json:"sublabel"`

    Validation struct {
        Id string `json:"id"`
        NumOfValid int32 `json:"num_yes"`
        NumOfInvalid int32 `json:"num_no"`
    }
}

type UnannotatedImage struct {
    Id string `json:"uuid"`
    Unlocked bool `json:"unlocked"`
    Url string `json:"url"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
    Provider string `json:"provider"`
    Width int32 `json:"width"`
    Height int32 `json:"height"`
    Validation struct {
        Id string `json:"uuid"`
    } `json:"validation"`
    AutoAnnotations []json.RawMessage `json:"auto_annotations,omitempty"`
}

type ImageLabel struct {
    Image struct {
        Id string `json:"uuid"`
        Unlocked bool `json:"unlocked"`
        Url string `json:"url"`
        Provider string `json:"provider"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Labels[] struct {
        Name string `json:"name"`
        Unlocked bool `json:"unlocked"`
        NumOfValid int32 `json:"num_yes"`
        NumOfInvalid int32 `json:"num_no"`
        Sublabels[] struct {
            Name string `json:"name"`
            NumOfValid int32 `json:"num_yes"`
            NumOfInvalid int32 `json:"num_no"`
        } `json:"sublabels"`
    } `json:"labels"`
}

type AnnotatedImage struct {
    Image struct {
        Id string `json:"uuid"`
        Unlocked bool `json:"unlocked"`
        Url string `json:"url"`
        Provider string `json:"provider"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Validation struct {
        Label string `json:"label"`
        Sublabel string `json:"sublabel"`
    } `json:"validation"`
    

    Id string `json:"uuid"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
    Annotations []json.RawMessage `json:"annotations"`
    NumRevisions int32 `json:"num_revisions"`
    Revision int32 `json:"revision"`
}

type ImageValidation struct {
    Uuid string `json:"uuid"`
    Valid string `json:"valid"`
}

type ImageValidationBatch struct {
    Validations []ImageValidation `json:"validations"`
}

type GraphNode struct {
	Group int `json:"group"`
	Text string `json:"text"`
	Size int `json:"size"`
}

type ValidationStat struct {
    Label string `json:"label"`
    Count int `json:"count"`
    ErrorRate float32 `json:"error_rate"`
    TotalValidations int `json:"total_validations"`
}

type DonationsPerCountryStat struct {
    CountryCode string `json:"country_code"`
    Count int64 `json:"num"`
}

type ValidationsPerCountryStat struct {
    CountryCode string `json:"country_code"`
    Count int64 `json:"num"`
}

type AnnotationsPerCountryStat struct {
    CountryCode string `json:"country_code"`
    Count int64 `json:"num"`
}

type DonationsPerAppStat struct {
    AppIdentifier string `json:"app_identifier"`
    Count int64 `json:"num"`
}

type ValidationsPerAppStat struct {
    AppIdentifier string `json:"app_identifier"`
    Count int64 `json:"num"`
}

type AnnotationsPerAppStat struct {
    AppIdentifier string `json:"app_identifier"`
    Count int64 `json:"num"`
}

type AnnotatedStat struct {
    Label struct {
        Id string `json:"uuid"`
        Name string `json:"name"`
    } `json:"label"`
    Num struct {
        Completed int64 `json:"completed"`
        Total int64 `json:"total"`
    } `json:"num"`
}

type Statistics struct {
    Validations []ValidationStat `json:"validations"`
    DonationsPerCountry []DonationsPerCountryStat `json:"donations_per_country"`
    ValidationsPerCountry []ValidationsPerCountryStat `json:"validations_per_country"`
    AnnotationsPerCountry []AnnotationsPerCountryStat `json:"annotations_per_country"`
    DonationsPerApp []DonationsPerAppStat `json:"donations_per_app"`
    ValidationsPerApp []ValidationsPerAppStat `json:"validations_per_app"`
    AnnotationsPerApp []AnnotationsPerAppStat `json:"annotations_per_app"`
    NumOfUnlabeledDonations int64 `json:"num_of_unlabeled_donations"`
}

type UnannotatedValidation struct {
    Validation struct {
        Id string `json:"uuid"`
        Label string `json:"label"`
        Sublabel string `json:"sublabel"`
    } `json:"validation"`
}

type LabelSearchItem struct {
    Label string `json:"label"`
    ParentLabel string `json:"parent_label"`
}

type LabelSearchResult struct {
    Labels []LabelSearchItem `json:"items"`
}

type AnnotationRefinementQuestion struct {
    Question string `json:"question"`
    Uuid int64 `json:"uuid"`
    RecommendedControl string `json:"recommended_control"`
}

type AnnotationRefinementAnswerExample struct {
    Filename string `json:"filename"`
    Attribution string `json:"attribution"`
}

type AnnotationRefinementAnswer struct {
    Label string `json:"label"`
    Id int64 `json:"id"`
    Examples []AnnotationRefinementAnswerExample `json:"examples"`
}

type AnnotationRefinement struct {
    Question AnnotationRefinementQuestion `json:"question"`
    //Answers []AnnotationRefinementAnswer `json:"answers"`
    Answers []json.RawMessage `json:"answers"`

    Metainfo struct {
        BrowseByExample bool `json:"browse_by_example"`
        AllowOther bool `json:"allow_other"`
        AllowUnknown bool `json:"allow_unknown"`
        MultiSelect bool `json:"multiselect"`
    } `json:"metainfo"`

    Image struct {
        Uuid string `json:"uuid"`
    } `json:"image"`

    Annotation struct{
        Uuid string `json:"uuid"`
        Annotation json.RawMessage `json:"annotation"`
    } `json:"annotation"`
}

type AnnotationRefinementEntry struct {
    LabelId string `json:"label_uuid"`
}

type BatchAnnotationRefinementEntry struct {
    LabelId string `json:"label_uuid"`
    AnnotationDataId string `json:"annotation_data_uuid"`
}

type ExportedImage struct {
    Id string `json:"uuid"`
    Provider string `json:"provider"`
    Width int32 `json:"width"`
    Height int32 `json:"height"`
    Annotations []json.RawMessage `json:"annotations"`
    Validations []json.RawMessage `json:"validations"`
}

type AutoAnnotationImage struct {
    Image struct {
        Id string `json:"uuid"`
        Provider string `json:"provider"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Labels []string `json:"labels"`
}


type UserStatistics struct {
    Total struct {
        Validations int32 `json:"validations"`
        Annotations int32 `json:"annotations"`
    } `json:"total"`

    User struct {
        Validations int32 `json:"validations"`
        Annotations int32 `json:"annotations"`
    } `json:"user"`
}

type UserPermissions struct {
    CanRemoveLabel bool `json:"can_remove_label"`
}

type UserInfo struct {
    Name string `json:"name"`
    Created int64 `json:"created"`
    ProfilePicture string `json:"profile_picture"`
    IsModerator bool `json:"is_moderator"`

    Permissions *UserPermissions `json:"permissions,omitempty"`
}

type DataPoint struct {
    Value int32 `json:"value"`
    Date string `json:"date"`
}

type Activity struct {
    Name string `json:"name"`
    Type string `json:"type"`
    Date string `json:"date"`
    Image struct {
        Id string `json:"uuid"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
        Annotation json.RawMessage `json:"annotation"`
        Label string `json:"label"`
    } `json:"image"`
}

type APIToken struct {
    IssuedAtUnixTimestamp int64 `json:"issued_at"`
    Token string `json:"token"`
    Description string `json:"description"`
    Revoked bool `json:"revoked"`
}

type AnnotationTask struct {
    Image struct {
        Id string `json:"uuid"`
        Unlocked bool `json:"unlocked"`
        Url string `json:"url"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Id string `json:"uuid"`
}

type AnnotationRefinementTask struct {
    Image struct {
        Id string `json:"uuid"`
        Unlocked bool `json:"unlocked"`
        Url string `json:"url"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Annotation struct{
        Id string `json:"uuid"`
        Data json.RawMessage `json:"data"`
    } `json:"annotation"`

    Refinements json.RawMessage `json:"refinements"`
}

type Report struct {
    Reason string `json:"reason"`
}

type Label struct {
    Name string `json:"name"`
}

type LabelMeValidation struct {
    Uuid string `json:"uuid"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"` 
}

type Sublabel struct {
    Name string `json:"name"`
    Unlocked bool `json:"unlocked"`
    Annotatable bool `json:"annotatable"`
    Uuid string `json:"uuid"`
    Validation *LabelMeValidation `json:"validation,omitempty"`
}

type LabelMeEntry struct {
    Label string `json:"label"` 
    Unlocked bool `json:"unlocked"` 
    Annotatable bool `json:"annotatable"` 
    Uuid string `json:"uuid"`
    Validation *LabelMeValidation `json:"validation,omitempty"`
    Sublabels []Sublabel `json:"sublabels"`
}


type ContributionsPerCountryRequest struct {
    CountryCode string `json:"country_code"`
    Type string `json:"type"`
}

type ContributionsPerAppRequest struct {
    AppIdentifier string `json:"app_identifier"`
    Type string `json:"type"`
}


type MetaLabelMapEntry struct {
    Description string  `json:"description"`
    Name string `json:"name"`
}

type LabelMapQuizExampleEntry struct {
    Filename string `json:"filename"`
    Attribution string `json:"attribution"`
}

type LabelMapQuizAnswerEntry struct {
    Name string `json:"name"`
    Examples []LabelMapQuizExampleEntry `json:"examples"`
    Uuid string `json:"uuid"`
}


type LabelMapQuizEntry struct {
    Question string `json:"question"`
    Accessors []string `json:"accessors"`
    Answers []LabelMapQuizAnswerEntry `json:"answers"`
    AllowUnknown bool `json:"allow_unknown"`
    AllowOther bool `json:"allow_other"`
    BrowseByExample bool `json:"browse_by_example"`
    Multiselect bool `json:"multiselect"`
    ControlType string `json:"control_type"`
    Uuid string `json:"uuid"`
}

type LabelMapEntry struct {
    Description string  `json:"description"`
    LabelMapEntries map[string]LabelMapEntry  `json:"has"`
    Accessors []string `json:"accessors"`
    Quiz []LabelMapQuizEntry `json:"quiz"`
    Uuid string `json:"uuid"`
}

type LabelMap struct {
    LabelMapEntries map[string]LabelMapEntry `json:"labels"`
    MetaLabelMapEntries map[string]MetaLabelMapEntry  `json:"metalabels"`
}

type LabelValidationEntry struct {
    Label string  `json:"label"`
    Sublabel string `json:"sublabel"`
}

type BlogSubscribeRequest struct {
    Email string `json:"email"`
}

type ImageSource struct {
    Provider string
    Url string
    Trusted bool
}

type ImageInfo struct {
    Hash uint64
    Width int32
    Height int32
    Name string
    Source ImageSource

}

type UserSignupRequest struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password"`
}

type APIUser struct {
    Name string `json:"name"`
    ClientFingerprint string `json:"client_fingerprint"`
}

type LabelMapRefinementValue struct {
    Uuid string `json:"uuid"`
    Description string `json:"description"`
    Accessors []string `json:"accessors"`
}

type LabelMapRefinementEntry struct {
     Values map[string]LabelMapRefinementValue `json:"values"`
     Uuid string `json:"uuid"`
     Icon string `json:"icon"`
}

type UpdateAnnotationCoverageRequest struct {
    Uuid string `json:"uuid"`
    Type string `json:"type"`
}